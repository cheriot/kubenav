package kubenav.models.k8s

import io.k8s.apimachinery.pkg.apis.meta.v1.LabelSelector
import io.k8s.api.core.v1.Service
import io.k8s.api.apps.v1.Deployment
import kubenav.models.k8s.K8sError._
import io.k8s.apimachinery.pkg.apis.meta.v1.LabelSelectorRequirement

trait HasPodSelector {
  def podSelectors: Option[LabelSelector]
}

object HasPodSelector {

  import scala.language.implicitConversions

  def matchPod(selector: LabelSelector, podLabels: Map[String, String]): Boolean = {
    // everything gets ANDed together (see comments in LabelSelector)
    val labelMatch = selector.matchLabels.map { requiredLabels =>
      requiredLabels.toSet.subsetOf(podLabels.toSet)
    } getOrElse true

    val exprMatch = selector.matchExpressions.map { matchExprs =>
      matchExprs
        .map { matchExpr =>
          evalMatchExpression(matchExpr, podLabels)
        }
        .fold(true)(_ && _)
    } getOrElse true

    // Guessing that an empty LabelSelector is impossible or matches everything
    labelMatch && exprMatch
  }

  def evalMatchExpression(matchExpr: LabelSelectorRequirement, podLabels: Map[String, String]): Boolean = {
    def evalIn: Option[Boolean] =
      for {
        required <- matchExpr.values
        podLabelValue <- podLabels.get(matchExpr.key)
      } yield required.toSet.contains(podLabelValue)

    def evalExists: Boolean =
      podLabels.keySet.contains(matchExpr.key)

    // Valid operators are In, NotIn, Exists and DoesNotExist
    matchExpr.operator match {
      case "In"           => evalIn getOrElse false
      case "NotIn"        => evalIn.map(!_) getOrElse false
      case "Exists"       => evalExists
      case "DoesNotExist" => !evalExists
    }
  }

  implicit def servicePodSelector(svc: Service): HasPodSelector =
    new HasPodSelector {
      def podSelectors: Option[LabelSelector] = {
        for {
          spec <- svc.spec
          selector <- spec.selector
        } yield LabelSelector(matchExpressions = None, matchLabels = Some(selector))
      }
    }

  implicit def deploymentPodSelector(dep: Deployment): HasPodSelector =
    new HasPodSelector {
      def podSelectors: Option[LabelSelector] =
        dep.spec.map(_.selector)
    }

  val fail = OperationNotSupported(this) _
  val notFound = NotFound(this) _

  def dynamic(any: Any): Either[K8sError, LabelSelector] = {
    val hasE: Either[K8sError, HasPodSelector] = any match {
      case s: Service    => Right(s)
      case d: Deployment => Right(d)
      case _ => Left(fail(any))
    }
    hasE.flatMap(
      _.podSelectors.toRight(notFound(any))
    )
  }
}
