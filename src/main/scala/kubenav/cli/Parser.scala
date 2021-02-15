package kubenav.cli
import kubenav.KnEnv._
import kubenav.models.k8s.ResourceType
import zio.logging._

import java.io.File

object Parser {

  /**
   * Parse command line options. Prefer similarity to the options described in
   * kubectl options
   * kubectl --help
   * kubectl (describe|get) --help
   *
   * In the future, get closer to the kubectl api with pod/name-123
   */
  def parse(args: List[String]): CommandLineParams = {

    val parser = new scopt.OptionParser[CommandLineParams](kubenav.BuildInfo.name) {
      head("Navigate k8s objects like a graph.", s"(${kubenav.BuildInfo.version})\n")

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
            .required()
            .action((v, o) => o.copy(namespace = v)),
          //opt[Map[String,String]]('l', "label")
          //  // TODO: validate https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
          //  .action((v, o) => o.copy(label = Some(v)))
          //  .text("key=value, the label to scope api-server queries"),
          opt[String]("relation")
            .text("a RESOURCE_TYPE that is related to the resource specified")
            .action((v, o) => o.copy(relation = ResourceType(v)))
            .validate { v =>
              ResourceType(v).toRight(s"Unsupported relation type $v").map(_ => ())
            },
          arg[String]("RESOURCE_TYPE")
            .hidden()
            .action((v, o) => o.copy(resourceType = ResourceType(v).get))
            .validate { v =>
              ResourceType(v).toRight(s"Unsupported RESOURCE_TYPE $v").map(_ => ())
            },
          arg[String]("RESOURCE_NAME").hidden().action((v, o) => o.copy(resourceName = v)),
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
