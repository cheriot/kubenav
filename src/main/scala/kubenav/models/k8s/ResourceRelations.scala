package kubenav.models.k8s

import kubenav.models.k8s.HasPodSelector
import kubenav.models.k8s.K8sError
import kubenav.models.k8s.PodLike
import kubenav.models.k8s.ResourceType._

object ResourceRelations {

  type RelationFilter = (Any, List[Any]) => Either[K8sError, List[Any]]

  lazy val known: Map[ResourceType, Map[ResourceType, RelationFilter]] =
    Map(
      Service -> Map(
        Deployment -> podSelectorRelation,
        ReplicaSet -> podSelectorRelation,
        Pod -> podSelectorRelation,
      ),
      Pod -> Map(
        ReplicaSet -> ownerRelation
      ),
      ReplicaSet -> Map(
        Deployment -> ownerRelation
      ),
    )

  // pod -> rs -> deployment
  // HasOwnerReference{apiVersion,kind,name,uid}
  // HasId{apiVersion,kind,name,namespace,uid}
  val ownerRelation: RelationFilter = { (origin, others) =>
    HasOwner.dynamic(origin).map { ownerUids =>
      val ownerSet = ownerUids.toNes

      others.filter { other =>
        val idE = HasIdentity.dynamic(other)
        idE.fold(_ => false, uid => ownerSet.contains(uid))
      }
    }
  }

  val podSelectorRelation: RelationFilter = { (origin, others) =>
    HasPodSelector.dynamic(origin).map { labelSector =>
      others
        .map { other =>
          PodLike.dynamic(other).map { podLike =>
            for {
              podMeta <- podLike.podMeta
              labels <- podMeta.labels
            } yield (other, HasPodSelector.matchPod(labelSector, labels))
          }
        }
        .collect {
          case Right(Some((other, isMatch))) if isMatch => other
        }
    }
  }

}
