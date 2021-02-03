package kubenav

import zio._

package object kube {
  type KubeRepo = Has[KubeRepo.Service]
}
