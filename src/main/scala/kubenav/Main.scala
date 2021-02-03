package kubenav

import io.chrisdavenport.log4cats
import zio._
import zio.console._
import zio.logging._
import zio.logging.slf4j._
import com.goyeau.kubernetes.client._
import io.k8s.api.core.v1.NamespaceList
import io.k8s.api.core.v1.Namespace

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
  import org.rogach.scallop._

  class CommandLineArgs(args: Seq[String]) extends ScallopConf(args) {
    val logLevel = choice(
      default = Some(defaultLogLevel.render),
      choices = logLevels.map(_.render),
      descr = "Log level of kubenav client."
    )
  }

  val defaultLogLevel = LogLevel.Error
  val logLevels = List(
    LogLevel.Off,
    LogLevel.Trace,
    LogLevel.Debug,
    LogLevel.Info,
    LogLevel.Warn,
    LogLevel.Error,
    LogLevel.Fatal
  )
  def matchLogLevel(levelName: String): Option[LogLevel] =
    logLevels.find(_.render == levelName)

  override def run(args: List[String]): ZIO[ZEnv, Nothing, zio.ExitCode] = {
    val parsedArgs = new CommandLineArgs(args)
    parsedArgs.verify()

    val logLevel = parsedArgs.logLevel.toOption
      .flatMap(matchLogLevel)
      // Ignore this error. Scallop will validate and default the user provided levelName
      .getOrElse(defaultLogLevel)

    val logging = Logging.consoleErr(
      logLevel = logLevel,
      format = LogFormat.ColoredLogFormat()
    ) >>> Logging.withRootLoggerName("kubenav-cli")

    val env = logging >>> KubeRepo.live

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
