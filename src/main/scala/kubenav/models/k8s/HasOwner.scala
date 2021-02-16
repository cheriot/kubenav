package kubenav.models.k8s

import kubenav.models.k8s.K8sError._
import io.k8s.api.core.v1.Pod
import io.k8s.api.apps.v1.ReplicaSet
import cats.data.NonEmptyList

trait HasOwner {
  def ownerUid: List[Uid]
}

object HasOwner {
  import scala.language.implicitConversions

  implicit def podHasOwner(p: Pod): HasOwner =
    new HasOwner {
      def ownerUid = {
        val uids = for {
          meta <- p.metadata
          ownerRefs <- meta.ownerReferences
        } yield ownerRefs.map(_.uid).map(Uid.apply).toList
        uids.toList.flatten
      }
    }

  implicit def replicaSetHasOwner(rs: ReplicaSet): HasOwner =
    new HasOwner {
      def ownerUid = {
        val uids = for {
          meta <- rs.metadata
          ownerRefs <- meta.ownerReferences
        } yield ownerRefs.map(_.uid).map(Uid.apply).toList
        uids.toList.flatten
      }
    }

  val fail = OperationNotSupported(this) _
  val notFound = NotFound(this) _

  def dynamic(any: Any): Either[K8sError, NonEmptyList[Uid]] = {
    any match {
      case rs: ReplicaSet => NonEmptyList.fromList(rs.ownerUid).toRight(notFound(rs))
      case p: Pod         => NonEmptyList.fromList(p.ownerUid).toRight(notFound(p))
      case _              => Left(fail(any))
    }
  }
}
