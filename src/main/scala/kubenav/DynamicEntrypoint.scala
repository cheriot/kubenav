package kubenav

import com.goyeau.kubernetes.client._
import kubenav.models.k8s.ResourceType
import zio._
import zio.logging._

object DynamicEntrypoint {
  import kubenav.models.k8s.K8sError
  import kubenav.models.k8s.K8sError._
  import kubenav.models.k8s.ResourceRelations.RelationFilter

  def get(
    client: KubernetesClient[Task],
    namespace: String,
    resourceType: ResourceType,
    resourceName: String,
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
    resourceType: ResourceType,
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
    relationFilter: RelationFilter,
  ): ZIO[Logging, K8sError, List[Any]] = {

    // Future enhancements:
    // * allow for a relationship to specify labels to filter the list query
    // * support for non-namespaced objects
    // * what are examples of objects that have relationships across namespaces?
    val queries = get(client, namespace, resourceType, resourceName) <&> list(client, namespace, relationType)
    val filtered = queries map { case (resource, relations) =>
      relationFilter(resource, relations)
    }

    filtered.either.map(_.flatten).absolve
  }

}
