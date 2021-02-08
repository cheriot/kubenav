package kubenav.cli

import kubenav.KnEnv._
import zio.logging._

import java.io.File

case class CommandLineParams(
  logLevel: LogLevel,
  kubeconfig: File,
  namespace: String,
  resourceType: String,
  resourceName: String,
  label: Option[Map[String, String]]
)
object CommandLineParams {
  def apply(): CommandLineParams =
    CommandLineParams(
      logLevel = defaultLogLevel,
      kubeconfig = new File(s"${System.getProperty("user.home")}/.kube/config"),
      namespace = "",
      label = None,
      resourceType = "",
      resourceName = ""
    )
}
