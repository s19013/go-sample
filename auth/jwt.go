package auth

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/s19013/go-sample/clock"
	"github.com/s19013/go-sample/entity"
)

const (
	RoleKey     = "role"
	UserNameKey = "user_name"
)

// Goの embed で鍵ファイルをバイナリに埋め込む
// これによりファイルを別途読み込む必要なし

//go:embed cert/secret.pem
var rawPrivKey []byte

//go:embed cert/public.pem
var rawPubKey []byte

// JWTの作成・検証を行うメインクラス
type JWTer struct {
	PrivateKey, PublicKey jwk.Key
	Store                 Store
	Clocker               clock.Clocker
}

// JWTの「ID（jti）」とユーザーIDを紐付けて保存
// セッション管理できるJWT (JWT + Redis のハイブリッド)
// これにより
// 強制ログアウトできる
// トークン無効化できる

//go:generate go run github.com/matryer/moq -out moq_test.go . Store
type Store interface {
	Save(ctx context.Context, key string, userID entity.UserID) error
	Load(ctx context.Context, key string) (entity.UserID, error)
}

func NewJWTer(s Store, c clock.Clocker) (*JWTer, error) {
	j := &JWTer{Store: s}

	privkey, err := parse(rawPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed in NewJWTer: private key: %w", err)
	}

	pubkey, err := parse(rawPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed in NewJWTer: public key: %w", err)
	}

	j.PrivateKey = privkey
	j.PublicKey = pubkey
	j.Clocker = c

	return j, nil
}

func parse(rawKey []byte) (jwk.Key, error) {
	key, err := jwk.ParseKey(rawKey, jwk.WithPEM(true))
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (j *JWTer) GenerateToken(ctx context.Context, u entity.User) ([]byte, error) {

	// JwtID() → トークンID（uuid）
	// Issuer() → 発行者
	// Subject() → 種類
	// IssuedAt() → 発行時間
	// Expiration() → 有効期限（30分）
	// Claim() → カスタム情報

	// role（権限）
	// user_name
	// を追加で入れている

	tok, err := jwt.NewBuilder().
		JwtID(uuid.New().String()).
		Issuer(`github.com/s19013/go-sample`).
		Subject("access_token").
		IssuedAt(j.Clocker.Now()).
		// redisのexpireはこれを使う。
		// https://pkg.go.dev/github.com/go-redis/redis/v8#Client.Set
		// clock.Durationだから Subする必要がある
		Expiration(j.Clocker.Now().Add(30*time.Minute)).
		Claim(RoleKey, u.Role).
		Claim(UserNameKey, u.Name).
		Build()

	if err != nil {
		return nil, fmt.Errorf("GenerateToken: failed to build token: %w", err)
	}

	// Storeに保存
	if err := j.Store.Save(ctx, tok.JwtID(), u.ID); err != nil {
		return nil, err
	}

	// RS256（公開鍵暗号）で署名
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, j.PrivateKey))
	if err != nil {
		return nil, err
	}
	return signed, nil
}

// JWT取得＆検証
func (j *JWTer) GetToken(ctx context.Context, r *http.Request) (jwt.Token, error) {
	// リクエストからJWT取得
	// Authorizationヘッダからトークンを取り出す
	token, err := jwt.ParseRequest(
		r,
		jwt.WithKey(jwa.RS256, j.PublicKey), // 署名検証に使う鍵を指定
		jwt.WithValidate(false),             // 今ここではバリデーションをしない(後でやってる)
	)

	if err != nil {
		return nil, err
	}

	// 有効期限チェックなど
	if err := jwt.Validate(token, jwt.WithClock(j.Clocker)); err != nil {
		return nil, fmt.Errorf("GetToken: failed to validate token: %w", err)
	}

	// Redisから削除して手動でexpireさせていることもありうる。
	// Redisに存在しない → 無効 → 強制ログアウト対応
	if _, err := j.Store.Load(ctx, token.JwtID()); err != nil {
		return nil, fmt.Errorf("GetToken: %q expired: %w", token.JwtID(), err)
	}
	return token, nil
}

// context.WithValue のキーとして使う専用の型
// context.WithValue(ctx, "user_id", uid) のように
// キーを直接書くと他のパッケージとキーが衝突する可能性あり、"user_id" が被ると上書きされる

// 一意な型になり衝突しない、外部から使えないため安心
type userIDKey struct{}
type roleKey struct{}

func (j *JWTer) FillContext(r *http.Request) (*http.Request, error) {
	token, err := j.GetToken(r.Context(), r)
	if err != nil {
		return nil, err
	}

	// userID取得
	uid, err := j.Store.Load(r.Context(), token.JwtID())
	if err != nil {
		return nil, err
	}

	// 情報をcontextに入れる
	ctx := SetUserID(r.Context(), uid)
	ctx = SetRole(ctx, token)
	clone := r.Clone(ctx)

	return clone, nil
}

func SetUserID(ctx context.Context, uid entity.UserID) context.Context {
	return context.WithValue(ctx, userIDKey{}, uid)
}

func GetUserID(ctx context.Context) (entity.UserID, bool) {
	id, ok := ctx.Value(userIDKey{}).(entity.UserID)
	return id, ok
}

func SetRole(ctx context.Context, tok jwt.Token) context.Context {
	get, ok := tok.Get(RoleKey)
	if !ok {
		return context.WithValue(ctx, roleKey{}, "")
	}
	return context.WithValue(ctx, roleKey{}, get)
}

func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey{}).(string)
	return role, ok
}

func IsAdmin(ctx context.Context) bool {
	role, ok := GetRole(ctx)
	if !ok {
		return false
	}
	return role == "admin"
}

// 全体の流れまとめ
// ① ログイン
// → GenerateToken
// → JWT発行 + Redis保存

// ② APIアクセス
// → GetToken
// → JWT検証 + Redis確認

// ③ contextに詰める
// → FillContext

// ④ handlerで使う
// → GetUserID, IsAdmin
