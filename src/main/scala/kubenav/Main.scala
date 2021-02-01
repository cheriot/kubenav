package kubenav

import io.chrisdavenport.log4cats
import zio._
import zio.logging._
import com.goyeau.kubernetes.client._
import io.k8s.api.core.v1.NamespaceList

package object kubenav {
  implicit def zioCatsLogger(implicit
      zlog: zio.logging.Logger[String]
  ) = new log4cats.Logger[Task] {
    def error(message: => String): Task[Unit] = zlog.error(message)
    def warn(message: => String): Task[Unit] = zlog.warn(message)
    def info(message: => String): Task[Unit] = zlog.info(message)
    def debug(message: => String): Task[Unit] = zlog.debug(message)
    def trace(message: => String): Task[Unit] = zlog.trace(message)
    def error(t: Throwable)(message: => String): Task[Unit] =
      zlog.locally(LogAnnotation.Throwable(Some(t))) {
        zlog.error(message)
      }
    def warn(t: Throwable)(message: => String): Task[Unit] =
      zlog.locally(LogAnnotation.Throwable(Some(t))) {
        zlog.warn(message)
      }
    def info(t: Throwable)(message: => String): Task[Unit] =
      zlog.locally(LogAnnotation.Throwable(Some(t))) {
        zlog.info(message)
      }
    def debug(t: Throwable)(message: => String): Task[Unit] =
      zlog.locally(LogAnnotation.Throwable(Some(t))) {
        zlog.debug(message)
      }
    def trace(t: Throwable)(message: => String): Task[Unit] =
      zlog.locally(LogAnnotation.Throwable(Some(t))) {
        zlog.trace(message)
      }
  }

  type KubeRepo = Has[KubeRepo.Service]
  object KubeRepo {
    trait Service {
      def use[A](f: KubernetesClient[Task] => Task[A]): ZIO[Any, Throwable, A]
    }

    val live: ZLayer[Logging, Nothing, KubeRepo] = ZLayer.fromService {
      implicit logging =>
        new Service {
          override def use[A](
              f: KubernetesClient[Task] => Task[A]
          ): ZIO[Any, Throwable, A] = {
            import java.io.File
            import zio.interop.catz._
            Task.concurrentEffectWith { implicit CE =>
              val kubeClient = KubernetesClient[Task](
                KubeConfig.apply[Task](
                  new File(s"${System.getProperty("user.home")}/.kube/config")
                )
              )
              kubeClient.use(f)
            }
          }
        }
    }

    def use[A](
        f: KubernetesClient[Task] => Task[A]
    ): ZIO[KubeRepo, Throwable, A] = {
      ZIO.accessM(_.get.use(f))
    }
  }
}

object Main extends zio.App {
  import kubenav._

  val env =
    Logging.consoleErr()

  override def run(args: List[String]): ZIO[ZEnv, Nothing, zio.ExitCode] =
    namespaceList
      .fold(_ => zio.ExitCode.failure, _ => zio.ExitCode.success)
      .provideLayer(KubeRepo.live)
      .provideLayer(env)

  def namespaceList: ZIO[KubeRepo, Throwable, NamespaceList] =
    KubeRepo.use[NamespaceList] { client =>
      client.namespaces.list
    }
}
