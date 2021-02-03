package kubenav.kube
import io.chrisdavenport.log4cats
import zio._
import zio.logging._
import com.goyeau.kubernetes.client._

object KubeClient {
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

  trait Service {
    def use[A](f: KubernetesClient[Task] => Task[A]): ZIO[Any, Throwable, A]
  }

  val live: ZLayer[Logging, Nothing, KubeClient] = ZLayer.fromService { implicit logging =>
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
  ): ZIO[KubeClient, Throwable, A] = {
    ZIO.accessM(_.get.use(f))
  }
}
