package kubenav.models.k8s

import io.k8s.api.core.v1.PodSpec
import io.k8s.api.apps.v1.ReplicaSet
import io.k8s.api.core.v1.Pod
import io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta
import io.k8s.apimachinery.pkg.apis.meta.v1.{LabelSelector, LabelSelectorRequirement}
import io.k8s.api.core.v1.Service
import io.k8s.api.core.v1.ServiceList
import io.k8s.api.apps.v1.Deployment
import io.k8s.api.apps.v1.DeploymentList
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
