package kubenav

import zio._

package object kube {
  type KubeClient = Has[KubeClient.Service]
}
