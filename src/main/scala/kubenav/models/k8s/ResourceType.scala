package kubenav.models.k8s

import cats.syntax.option._

sealed trait ResourceType

object ResourceType {
  // Initial subset of supported resources
  // kubectl api-resources -o wide
  case object Namespace extends ResourceType
  case object Service extends ResourceType
  case object Deployment extends ResourceType
  case object ReplicaSet extends ResourceType
  case object Pod extends ResourceType

  def apply(name: String): Option[ResourceType] =
    name match {
      case "namespace" | "ns"  => Namespace.some
      case "service"           => Service.some
      case "deployment"        => Deployment.some
      case "replicaset" | "rs" => ReplicaSet.some
      case "pod" | "po"        => Pod.some
      case _                   => None
    }
}
