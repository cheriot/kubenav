package kubenav

import zio._
import zio.logging._
import com.goyeau.kubernetes.client._
import io.k8s.apimachinery.pkg.apis.meta.v1.{LabelSelector, LabelSelectorRequirement}
import io.k8s.api.core.v1.Service
import io.k8s.api.core.v1.ServiceList
import io.k8s.api.apps.v1.Deployment
import io.k8s.api.apps.v1.DeploymentList
import kube.KubeClient
import kubenav.models.k8s.ResourceType
import io.k8s.api.core.v1.PodSpec
import io.k8s.api.apps.v1.ReplicaSet
import io.k8s.api.core.v1.Pod
import io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta

object DynamicEntrypoint {
  import kubenav.models.k8s.K8sError
  import kubenav.models.k8s.K8sError._
  import kubenav.models.k8s.ResourceRelations
  import kubenav.models.k8s.ResourceRelations.RelationFilter

  def get(
    client: KubernetesClient[Task],
    namespace: String,
    resourceType: ResourceType,
    resourceName: String
  ): ZIO[Logging, K8sError, Any] = {

    val description = s"get $namespace $resourceType $resourceName"
    log.trace(description) *>
      DynamicClient
        .get(client, namespace, resourceType, resourceName)
        .mapError(t => ClientException(t, s"Error attempting to $description"))
        .tapError { error =>
          log.throwable(error.msg, error.t)
        }
  }

  def list(
    client: KubernetesClient[Task],
    namespace: String,
    resourceType: ResourceType
  ): ZIO[Logging, K8sError, List[Any]] = {

    import zio.interop.catz._

    val description = s"list $namespace $resourceType"
    log.trace(description) *>
      DynamicClient
        .list(client, namespace, resourceType)
        .mapError { t =>
          ClientException(t, s"Error attempting to $description")
        }
        .tapError { error =>
          log.throwable(error.msg, error.t)
        }
        .tap(v => log.trace(s"list $namespace $resourceType ${v.size} result(s)"))
  }


  def relation(
    client: KubernetesClient[Task],
    namespace: String,
    resourceType: ResourceType,
    resourceName: String,
    relationType: ResourceType,
    relationFilter: RelationFilter
  ): ZIO[Logging, K8sError, List[Any]] = {

    val queries = get(client, namespace, resourceType, resourceName) <&> list(client, namespace, relationType)
    val filtered = queries map { case (resource, relations) =>
      relationFilter(resource, relations)
    }

    filtered.either.map(_.flatten).absolve
  }

}
