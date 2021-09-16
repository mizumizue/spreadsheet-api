# GoogleSpreadSheetを簡易APIにする

## 手順

- ServiceAccount作成
- SpreadSheetの共有ユーザーとしてCloudRunのサービスアカウントを追加する
– CloudRunへデプロイ(認証付き)
- CloudRun認証用トークン取得
- Request

### ServiceAccount作成

SpreadSheetを共有するServiceAccountを作成する

### SpreadSheetの共有ユーザーとしてCloudRunのサービスアカウントを追加する

SpreadSheetを共有する

### CloudRunへデプロイ(認証付き)

`make run-deploy`

※Makefileを編集し、サービス名他環境に合わせた値にすること

### CloudRun認証用トークン取得

`gcloud auth print-identity-token`

### Request

`{TOKEN}` には前の手順で取得したトークンを利用する

sheetIdとsheetNameパラメータは必須

それぞれSpreadSheetのIDとシート名を渡す

```
curl --location --request GET '{CLOUD_RUN_URL}/api/resources?sheetId={SPREAD_SHEET_ID}&sheetName={SPREAD_SHEET_NAME}' \
--header 'Content-Type: application/json; charset=utf-8' \
--header 'Authorization: Bearer {TOKEN}
```
