package kubenav.models.k8s

import io.k8s.apimachinery.pkg.apis.meta.v1.{LabelSelector, LabelSelectorRequirement}
import io.k8s.api.core.v1.Service
import io.k8s.api.core.v1.ServiceList
import io.k8s.api.apps.v1.Deployment
import io.k8s.api.apps.v1.DeploymentList
import io.k8s.api.core.v1.PodSpec
import io.k8s.api.apps.v1.ReplicaSet
import io.k8s.api.core.v1.Pod
import io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta
import kubenav.models.k8s.K8sError
import kubenav.models.k8s.K8sError._

trait PodLike {
  def podMeta: Option[ObjectMeta]
  def podSpec: Option[PodSpec]
}

object PodLike {
  import scala.language.implicitConversions

  implicit def deploymentPodSpec(d: Deployment): PodLike =
    new PodLike {
      def podMeta =
        for {
          dSpec <- d.spec
          meta <- dSpec.template.metadata
        } yield meta

      def podSpec =
        for {
          dSpec <- d.spec
          podSpec <- dSpec.template.spec
        } yield podSpec
    }

  implicit def replicaSetPodSpec(rs: ReplicaSet): PodLike =
    new PodLike {
      def podMeta =
        for {
          rsSpec <- rs.spec
          podTemplate <- rsSpec.template
          meta <- podTemplate.metadata
        } yield meta

      def podSpec =
        for {
          rsSpec <- rs.spec
          podTemplate <- rsSpec.template
          podSpec <- podTemplate.spec
        } yield podSpec
    }

  implicit def podPodSpec(p: Pod): PodLike =
    new PodLike {
      def podMeta = p.metadata
      def podSpec = p.spec
    }

  val fail = OperationNotSupported(this) _
  def dynamic(any: Any): Either[K8sError, PodLike] =
    any match {
      case d: Deployment => Right(d)
      case rs: ReplicaSet => Right(rs)
      case p: Pod => Right(p)
      case _          => Left(fail(any))
    }
}
