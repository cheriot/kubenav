package com.goyeau.kubernetes.client

import cats._
import cats.syntax.functor._
import kubenav.models.k8s.ResourceType
import kubenav.models.k8s.ResourceType._

object DynamicClient {
  // In com.goyeau.kubernetes.client because the interfaces needed are package private.

  def get[F[_]](client: KubernetesClient[F], namespace: String, resourceType: ResourceType, resourceName: String): F[_] = {
    resourceType match {
      case Namespace  => client.namespaces.get(namespace)
      case Service    => client.services.namespace(namespace).get(resourceName)
      case Deployment => client.deployments.namespace(namespace).get(resourceName)
      case ReplicaSet => client.replicaSets.namespace(namespace).get(resourceName)
      case Pod        => client.pods.namespace(namespace).get(resourceName)
    }
  }

  def list[F[_]: Functor](client: KubernetesClient[F], namespace: String, resourceType: ResourceType): F[List[_]] = {
    resourceType match {
      case Namespace  => client.namespaces.list().map(_.items.toList)
      case Service    => client.services.namespace(namespace).list().map(_.items.toList)
      case Deployment => client.deployments.namespace(namespace).list().map(_.items.toList)
      case ReplicaSet => client.replicaSets.namespace(namespace).list().map(_.items.toList)
      case Pod        => client.pods.namespace(namespace).list().map(_.items.toList)
    }
  }
}
