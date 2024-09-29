<h1 align="center">
  DAIJAI - GO
</h1>

## DEPLOYMENT

### login gcloud

1. login `gcloud auth login`
2. set project id `gcloud config set project daijai`

### Build step

1. change environment (.env file) to prod
2. comment some lines in `.gitignore`
   - keys/\*
   - .env
   - .env\*
3. build `gcloud builds submit --config cloudbuild_run.yaml`
4. deploy `gcloud run deploy daijai --platform managed --region asia-southeast1 --image gcr.io/daijai/daijai-go --allow-unauthenticated`

### Test locally

- `docker build --platform linux/amd64 -t daijai-go-1 .`
- `PORT=8080 && docker run -p 9090:${PORT} -e PORT=${PORT} .`
- `docker run --entrypoint=sh -ti daijai-go`

### Config Slug

1. set slug config in `slug.model.go`
2. add model to function `initSlugger` in `migrator.go`
3. run migrate
