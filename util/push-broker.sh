#/bin/bash
if [ $# -eq 0 ]; then
    echo "No tag name provided. You must provide a tag name to push to AWS."
    exit 1
else
    echo "$1"
fi

echo "Building broker..."
docker build -t helgart .
echo "Finished building broker!"
docker tag helgart:latest $1
echo "Pushing image..."
docker push $1
echo "Push complete!"