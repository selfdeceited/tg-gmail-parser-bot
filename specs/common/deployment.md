## Deployment
1. Use Github Actions for deployment. Steps will be:
  - to install the dependencies and build Docker image
  - to publish the created Docker image to container registry - ghcr.io. It should have a tag with the commit SHA.
  - when published, if it's main branch, use the corresponding tag to deploy the image to the production environment. We will use fly.io
2. Have Devcontainer config for local development
3. (do not implement yet) Add test suite to check 'happy path' scenarios
