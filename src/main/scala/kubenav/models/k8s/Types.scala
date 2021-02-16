package kubenav.models.k8s

case class Uid(val underlying: String) extends AnyVal
object Uid {
  import scala.language.implicitConversions

  implicit val orderingUid: cats.Order[Uid] =
    new cats.Order[Uid] {
      override def compare(x: Uid, y: Uid): Int = x.underlying.compare(y.underlying)
    }
}
