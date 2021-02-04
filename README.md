## Develop
1. export JAVA_HOME="$(/usr/libexec/java_home -v "11")"
1. brew install sbt kind
1. make cluster-create run

## Native
1. brew install --cask graalvm/tap/graalvm-ce-java11
1. xattr -r -d com.apple.quarantine /Library/Java/JavaVirtualMachines/graalvm-ce-java11-21.0.0
1. /Library/Java/JavaVirtualMachines/graalvm-ce-java11-21.0.0/Contents/Home/bin/gu install native-image