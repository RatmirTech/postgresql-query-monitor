package s3

type S3Client interface {
	CreateBucketIfNotExists(bucket string) error
	PutObject(bucket, key string, stream []byte) error
	GetObject(bucket, key string) ([]byte, error)
	DeleteObject(bucket, key string) error
}
