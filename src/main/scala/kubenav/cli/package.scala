package kubenav

import zio.logging._

import java.io.File

import KnEnv._

package object cli {

  case class CommandLineParams(
    logLevel: LogLevel,
    kubeconfig: File,
    namespace: Option[String],
    label: Option[String],
    resourceType: Option[String],
    resourceName: Option[String]
  )
  object CommandLineParams {
    def apply(): CommandLineParams =
      CommandLineParams(
        logLevel = defaultLogLevel,
        kubeconfig = new File(s"${System.getProperty("user.home")}/.kube/config"),
        namespace = None,
        label = None,
        resourceType = None,
        resourceName = None
      )
  }

  /**
   * Parse command line options. Prefer similarity to the options described in
   * kubectl options
   * kubectl --help
   * kubectl (describe|get) --help
   *
   * In the future, get closer to the kubectl api with pod/name-123
   */
  def parse(args: List[String]): CommandLineParams = {

    val parser = new scopt.OptionParser[CommandLineParams](BuildInfo.name) {
      head("Navigate k8s objects like a graph.", s"(${BuildInfo.version})\n")

      implicit val logLevelRead: scopt.Read[zio.logging.LogLevel] =
        scopt.Read.reads { str =>
          logLevels.find(_.render == str) match {
            case Some(ll) => ll
            case None     => throw new IllegalArgumentException(s"$str is not a valid log level")
          }
        }

      cmd("open")
        .children(
          opt[String]('n', "namespace")
            // TODO: validate https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
            .text("the namespace to scope api-server queries")
            .action((v, o) => o.copy(namespace = Some(v))),
          //opt[String]('l', "label")
          //  // TODO: validate https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
          //  .action((v, o) => o.copy(label = Some(v)))
          //  .text("key=value, the label to scope api-server queries"),
          arg[String]("RESOURCE_TYPE")
            .hidden()
            .action((v, o) => o.copy(resourceType = Some(v)))
            .validate(v =>
              if (v == "service") Right(())
              else Left("Only 'service' resources are supported right now")
            ),
          arg[String]("RESOURCE_NAME").hidden().action((v, o) => o.copy(resourceName = Some(v)))
        )

      opt[LogLevel]("log-level")
        .action((v, o) => o.copy(logLevel = v))
        .text("One of error, warn, info, debug, trace")
      opt[File]("kubeconfig")
        .text("Path to the kubeconfig file to use. Defaults to ~/.kube/config")
      help("help").hidden()

    }

    parser.parse(args, CommandLineParams()) match {
      case Some(result) =>
        println(result)
        result
      case None =>
        sys.exit(1)
    }
  }

}
