<h1 align="center">
  DAIJAI - GO
</h1>

## TODO:

- [x] รับ quantity ทศนิยมได้
- [x] ทศนิยม 2 ตำแหน่ง
- [x] เพิ่ม process การเลือก order ที่จะผลิต
- [x] ตอนรับของ มีใส่ PR แล้ว suggest รายการของ PR มาให้เลือกได้ (suggest เหมือน PR)
- [x] sum ของ main-inventory , หน้างาน
- [ ] ของหน้างาน ให้รู้ว่าเบิกไปโครงการไหน รับของโครงการไหน
- [x] เพิ่ม form เบิกแบบเลือกเข้า inventory
- [x] ตอนรับของ มีเลือกคลัง ถ้าไม่ใช่คลังหลัก ไม่มีการตัดไปทำ order
- [x] เพิ่ม transfer ของ คลังต่อคลัง

### 19/07/2024

- [ ] Drawing ให้ล็อครหัส
- [ ] หน้า receipt ราคาไม่ได้หาร 100
- [ ] เพิ่มเลือก order ตอน admin withdraw
- [ ] แยก HOF กับ Factory
- [ ] เพิ่ม layer คลังหน้างาน
- [ ] สร้าง user ให้ลองใช้

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
