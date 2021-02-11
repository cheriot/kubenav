package kubenav

import io.k8s.api.core.v1.Service
import io.k8s.api.core.v1.ServiceList
import io.k8s.api.apps.v1.Deployment
import zio._
import zio.console._

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
  case class Relation(
    origin: String,
    destination: String
    // queryType:
  )
  val relations = Map(
    // "service" -> [

    // ]
  )

  def namespaceList(
    namespace: String,
    resourceType: String,
    resourceName: String,
    relation: Option[String]
  ): ZIO[KubeClient, Nothing, Either[String, List[String]]] =
    KubeClient
      .use[Either[String, List[String]]] { client =>
        val resource: ZIO[Any, String, Service] = client.services
          .namespace(namespace)
          .get(resourceName)
          .either
          .map(_.left.map(t => s"Error fetching $resourceType/$resourceName in $namespace"))
          .absolve

        val relatedResourceO = relation.map { relationType =>
          val podSelectorZ: ZIO[Any, String, Map[String, String]] = resource
            .map { service =>
              val l = for {
                spec <- service.spec
                labels <- spec.selector
              } yield labels
              l match {
                case Some(ls) if ls.nonEmpty =>
                  Right(ls)
                case _ =>
                  Left(
                    s"$resourceType $resourceName does not have pod selectors so we don't know how to find a related $relationType"
                  )
              }
            }
            .either
            .map(_.flatMap(identity))
            .absolve

          val deploymentWithPodLabelsZ: ZIO[Any, String, List[(Deployment, Map[String, String])]] =
            client.deployments
              .namespace(namespace)
              .list()
              .map { deploymentList =>
                deploymentList.items.map { deployment =>
                  (deployment, podTemplateLabels(deployment))
                }.toList
              }
              .mapError(t => s"Error in fetching $relationType in $namespace: ${t.getMessage}")

          for {
            podSelector <- podSelectorZ
            deploymentWithPodLabels <- deploymentWithPodLabelsZ
          } yield {
            deploymentWithPodLabels.collect {
              case (deployment, podLabels)
                  if podSelector.forall(kv => podLabels.toList.contains(kv)) =>
                deployment
            }.toList
          }
        }

        // Serialize
        val relatedStringsO = relatedResourceO.map { relatedResourceZ =>
          relatedResourceZ.map { relatedResource =>
            relatedResource.flatMap(describeK8sObject).toList
          }
        }
        val resourceStrings = resource.map(describeK8sObject)

        relatedStringsO.getOrElse(resourceStrings).either
      }
      .either
      .map(_.left.map(t => s"Unexpected error ${t.getMessage}").flatMap(identity))

  def describeK8sObject(service: Service): List[String] = {
    import io.circe.generic.auto._, io.circe.syntax._
    List(service.asJson.spaces4)
  }

  def describeK8sObject(deployment: Deployment): List[String] = {
    import io.circe.generic.auto._, io.circe.syntax._
    List(deployment.asJson.spaces4)
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
