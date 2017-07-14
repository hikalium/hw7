BLACK=HUMAN
WHITE=https://othello-greedy.appspot.com
#
URL=https://step-reversi.appspot.com
VIEWERPATH=`curl "$URL/new?black=$BLACK&white=$HUMAN" | sed -e 's/.*"\(.*\)".*/\1/g'`
VIEWERURL=$URL$VIEWERPATH
REFLECTOR_ARGS=`curl $VIEWERURL | grep go | sed -e 's/.*"\go run reflector.go \(.*\)".*/\1/g' | cut -d ' ' -f1,3`
go run reflector.go $REFLECTOR_ARGS
