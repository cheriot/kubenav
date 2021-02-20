package kubenav.models.k8s

import io.k8s.api.apps.v1.Deployment
import io.k8s.api.apps.v1.ReplicaSet
import io.k8s.api.core.v1.Pod
import kubenav.models.k8s.K8sError._

trait HasIdentity {
  def uid: Option[Uid]
}

object HasIdentity {
  import scala.language.implicitConversions

  implicit def deploymentHasId(d: Deployment): HasIdentity =
    new HasIdentity {
      def uid: Option[Uid] = d.metadata.flatMap(_.uid).map(Uid.apply)
    }

  implicit def replicaSetHasId(rs: ReplicaSet): HasIdentity =
    new HasIdentity {
      def uid: Option[Uid] = rs.metadata.flatMap(_.uid).map(Uid.apply)
    }

  implicit def podHasId(p: Pod): HasIdentity =
    new HasIdentity {
      def uid: Option[Uid] = p.metadata.flatMap(_.uid).map(Uid.apply)
    }

  val fail = OperationNotSupported(this) _
  val notFound = NotFound(this) _

  def dynamic(any: Any): Either[K8sError, Uid] = {
    any match {
      case d: Deployment  => d.uid.toRight(notFound(d))
      case rs: ReplicaSet => rs.uid.toRight(notFound(rs))
      case p: Pod         => p.uid.toRight(notFound(p))
      case _              => Left(fail(any))
    }
  }
}
