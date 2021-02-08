package kubenav

import io.k8s.api.core.v1.Service
import io.k8s.api.core.v1.ServiceList
import zio._
import zio.console._
import zio.logging._

import kube.KubeClient

object Main extends zio.App {

  override def run(args: List[String]): ZIO[ZEnv, Nothing, zio.ExitCode] = {
    val cliArgs = cli.Parser.parse(args)

    namespaceList(cliArgs.namespace, cliArgs.resourceType, cliArgs.resourceName)
      .flatMap { names =>
        putStrLn(names.mkString(", "))
      }
      .flatMapError { e =>
        log.throwable("Unable to find resource.", e)
      }
      .fold(_ => zio.ExitCode.failure, _ => zio.ExitCode.success)
      .provideSomeLayer(KnEnv.env(cliArgs))
  }

  // service#spec#selector{key:value}
  // replicaSet#ownerReferences[{kind,name}], replicaSet#metadata#labels
  // pod#metadata#ownerReferences[{kind,name}], pod#metadata#labels

  def namespaceList(
    namespace: String,
    resourceType: String,
    resourceName: String
  ): ZIO[KubeClient, Throwable, List[String]] =
    KubeClient
      .use[List[String]] { client =>
        client.services.namespace(namespace).get(resourceName).map(describeK8sObject)
      }

  def describeK8sObject(service: Service): List[String] = {
    import io.circe.generic.auto._, io.circe.syntax._

    List(service.asJson.spaces4)
  }

  def nameStrings(nsList: ServiceList): List[String] =
    nsList.items.map { n: Service =>
      n.metadata.flatMap(_.name).getOrElse("[unnamed]")
    }.toList
}
