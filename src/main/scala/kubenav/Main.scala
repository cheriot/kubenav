package kubenav

import zio._
import zio.console._
import zio.logging._
import kube.KubeRepo
import io.k8s.api.core.v1.NamespaceList
import io.k8s.api.core.v1.Namespace

object Main extends zio.App {

  override def run(args: List[String]): ZIO[ZEnv, Nothing, zio.ExitCode] = {
    val cliArgs = cli.parse(args)

    val env = KnEnv.loggingLayer(cliArgs.logLevel) >>> KubeRepo.live

    namespaceList
      .flatMap { names =>
        putStrLn(names.mkString(", "))
      }
      .fold(_ => zio.ExitCode.failure, _ => zio.ExitCode.success)
      .provideSomeLayer(env)
  }

  def namespaceList: ZIO[KubeRepo, Throwable, List[String]] =
    KubeRepo
      .use[List[String]] { client =>
        client.namespaces.list.map(nameStrings)
      }

  def nameStrings(nsList: NamespaceList): List[String] =
    nsList.items.map { n: Namespace =>
      n.metadata.flatMap(_.name).getOrElse("[unnamed]")
    }.toList
}
