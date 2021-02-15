package kubenav.cli

import kubenav.KnEnv._
import zio.logging._

import java.io.File
import kubenav.models.k8s.ResourceType

case class CommandLineParams(
  logLevel: LogLevel,
  kubeconfig: File,
  namespace: String,
  resourceType: ResourceType,
  resourceName: String,
  relation: Option[ResourceType],
  label: Option[Map[String, String]]
)
object CommandLineParams {
  def apply(): CommandLineParams =
    CommandLineParams(
      logLevel = defaultLogLevel,
      kubeconfig = new File(s"${System.getProperty("user.home")}/.kube/config"),
      namespace = "",
      resourceType = ResourceType.Pod,
      resourceName = "",
      relation = None,
      label = None
    )
}
