package kubenav
import io.k8s.api.apps.v1.Deployment
import io.k8s.api.apps.v1.DeploymentList
import io.k8s.api.core.v1.Service
import io.k8s.api.core.v1.ServiceList
import kubenav.models.k8s.K8sError._
import kubenav.models.k8s.ResourceRelations
import kubenav.models.k8s.ResourceType
import zio._
import zio.console._
import zio.logging._

import kube.KubeClient

object Main extends zio.App {

  override def run(args: List[String]): ZIO[ZEnv, Nothing, zio.ExitCode] = {
    val cliArgs = cli.Parser.parse(args)

    namespaceList(
      cliArgs.namespace,
      cliArgs.resourceType,
      cliArgs.resourceName,
      cliArgs.relation
    ).absolve
      .flatMap { (output: List[String]) =>
        putStrLn(("Success!" +: output).mkString("\n"))
      }
      .flatMapError { errMsg =>
        putStrLn(s"Error: $errMsg")
      }
      .fold(_ => zio.ExitCode.failure, _ => zio.ExitCode.success)
      .provideSomeLayer(KnEnv.env(cliArgs))
  }

  // service#spec#selector{matchLabels,matchExpression}
  // service only has relations IFF #spec#selector is non-empty
  // #spec#selector will ONLY match pods in the same namespace
  // service#spec#externalName maybe have a cluster DNS that links to something in the namespace

  // deployment will not have the labels from the service's selector
  // deployment#spec#selector{matchLabels,matchExpression}

  // replicaSet#ownerReferences[{kind,name}], replicaSet#metadata#labels

  // pod#metadata#ownerReferences[{kind,name}], pod#metadata#labels

  // def ListAndCompareSelector(resourceType: String, selector: Object): Unit

  object K8sTypes {
    /*
     * io.k8s types have a lot of convention that's not captured in the type level. These traits bring those shared abilities
     * into the type system.
     */

    trait K8sList[T] {
      def items: Seq[T]
      def apiVersion: Option[String]
      def kind: Option[String]
      def metadata: Option[io.k8s.apimachinery.pkg.apis.meta.v1.ListMeta]
    }
    object K8sList {
      import scala.language.implicitConversions

      implicit def deploymentK8sList(dl: DeploymentList): K8sList[Deployment] =
        new K8sList[Deployment] {
          def items: Seq[Deployment] = dl.items
          def apiVersion: Option[String] = dl.apiVersion
          def kind: Option[String] = dl.kind
          def metadata: Option[io.k8s.apimachinery.pkg.apis.meta.v1.ListMeta] = dl.metadata
        }

    }

    // Valid operators are In, NotIn, Exists and DoesNotExist
  }

  def namespaceList(
    namespace: String,
    resourceType: ResourceType,
    resourceName: String,
    relation: Option[ResourceType]
  ): ZIO[KubeClient with Logging, Nothing, Either[String, List[String]]] = {
    val serialized = KubeClient
      .use[Logging, Either[String, List[String]]] { client =>
        val result = for {
          targetRelation <- relation
          originRelations <- ResourceRelations.known.get(resourceType)
          filter <- originRelations.get(targetRelation)
        } yield {
          DynamicEntrypoint.relation(client, namespace, resourceType, resourceName, targetRelation, filter)
        }

        import cats.syntax.traverse._
        import zio.interop.catz._

        val flatZ = result
          .traverse(identity)
          .map { (r: Option[List[Any]]) =>
            r.toRight(UnknownRelation(resourceType, relation.get))
          }
          .either
          .map(_.flatMap(identity))
          .absolve

        // serialize
        flatZ
          .mapError(k8sError => s"$k8sError")
          .map(
            _.flatMap(describeK8sObject)
          )
          .either
      }

    serialized
      .flatMapError(t => log.throwable("Error using KubeClient", t).map(_ => "Error communicating with cluster"))
      .either
      .map(_.flatMap(identity))
  }

  def describeK8sObject(obj: Any): List[String] = {
    import io.circe.generic.auto._, io.circe.syntax._
    obj match {
      case s: Service    => List(s.asJson.spaces4)
      case d: Deployment => List(d.asJson.spaces4)
      case _             => List(s"Don't know how to print object $obj")
    }
  }

  def podTemplateLabels(d: Deployment): Map[String, String] = {
    val labels = for {
      spec <- d.spec
      metadata <- spec.template.metadata
      labels <- metadata.labels
    } yield labels

    labels.getOrElse(Map())
  }

  def nameStrings(nsList: ServiceList): List[String] =
    nsList.items.map { n: Service =>
      n.metadata.flatMap(_.name).getOrElse("[unnamed]")
    }.toList
}
