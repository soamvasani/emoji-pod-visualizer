
tag=$1
if [ -z "$tag" ]
then
    echo "Using latest tag"
    tag=latest
fi

GOOS=linux GOARCH=386 go build

docker build -t soam/emoji-pod-vis:$tag .
docker push soam/emoji-pod-vis:$tag 
kubectl set image deployment/emoji-pod-vis emoji-pod-vis=soam/emoji-pod-vis:$tag
