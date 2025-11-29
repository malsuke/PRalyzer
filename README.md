
## 実行時引数

 対象プロジェクトのGitHubリポジトリURL, GitHub PAT
 
## 処理の流れ

1. ワードリストから1単語取得する
2. リポジトリのURLをもとに検索を行う
3. 検索して取得したPRの番号を収集する
（取得した番号は配列に格納する）
4. 配列の中身の重複を削除する
5. issues/<PRの番号>/commentsエンドポイントを使ってコメントを取得する
6. <PRの番号>.jsonでPR内のコメントをファイルに書き込む

# ディレクトリ構成

- data/<owner>/<repo>/<PRの番号>.json

- internal/github/search_pull_requests.go githubのpull requestsを検索する
