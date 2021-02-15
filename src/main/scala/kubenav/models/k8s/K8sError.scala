package kubenav.models.k8s

sealed trait K8sError
object K8sError {
  case class ClientException(t: Throwable, msg: String) extends K8sError
  case class UnknownRelation(resourceType: ResourceType, relationType: ResourceType) extends K8sError
  case class RelationNotSupported(resource: Any) extends K8sError

  case class OperationNotSupported(op: String, param: String) extends K8sError
  object OperationNotSupported {
    def apply(op: Any)(param: Any): OperationNotSupported =
      OperationNotSupported(op.getClass().getCanonicalName(), param.getClass().getCanonicalName())
  }

  case class NotFound(op: String, param: String) extends K8sError
  object NotFound {
    def apply(op: Any)(param: Any): NotFound =
      NotFound(op.getClass().getCanonicalName(), param.getClass().getCanonicalName())
  }
}
