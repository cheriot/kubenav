package kubenav.models.k8s
import kubenav.models.k8s.HasPodSelector
import kubenav.models.k8s.K8sError
import kubenav.models.k8s.PodLike

object ResourceRelations {

  type RelationFilter = (Any, List[Any]) => Either[K8sError, List[Any]]

  lazy val known: Map[ResourceType, Map[ResourceType, RelationFilter]] =
    Map(
      ResourceType.Service -> Map(
        ResourceType.Deployment -> podSelectorRelation
      )
    )

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
