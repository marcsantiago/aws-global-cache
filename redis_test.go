package awsglobalcache

import (
	"io/ioutil"
	"testing"

	"github.com/go-redis/redis/v8"
)

func init() {
	WarningLogger.SetOutput(ioutil.Discard)
}

func TestConfiguration_RetrieveRedisClient(t *testing.T) {
	type fields struct {
		awsRegion AWSRegion
		writer    *redis.Client
		readers   mappedRedisRegions
	}
	type args struct {
		operation Operation
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *redis.Client
	}{
		{
			name: "should pick the writer because the operation was set to write",
			fields: fields{
				awsRegion: USEast1,
				writer: redis.NewClient(&redis.Options{
					Addr: "writer_client",
				}),
				readers: nil,
			},
			args: args{operation: Write},
			want: redis.NewClient(&redis.Options{
				Addr: "writer_client",
			}),
		},
		{
			name: "should pick the reader because the operation was set to read",
			fields: fields{
				awsRegion: USWest1,
				writer: redis.NewClient(&redis.Options{
					Addr: "writer_client",
				}),
				readers: mappedRedisRegions{
					USWest1: redis.NewClient(&redis.Options{
						Addr: "reader_client",
					}),
				},
			},
			args: args{operation: Read},
			want: redis.NewClient(&redis.Options{
				Addr: "reader_client",
			}),
		},
		{
			name: "should pick the reader in the correct location",
			fields: fields{
				awsRegion: USWest1,
				writer: redis.NewClient(&redis.Options{
					Addr: "writer_client",
				}),
				readers: mappedRedisRegions{
					USEast1: redis.NewClient(&redis.Options{
						Addr: "reader_client_east",
					}),
					USWest1: redis.NewClient(&redis.Options{
						Addr: "reader_client_west",
					}),
				},
			},
			args: args{operation: Read},
			want: redis.NewClient(&redis.Options{
				Addr: "reader_client_west",
			}),
		},
		{
			name: "should pick the writer as a default reader is there is a region mismatch",
			fields: fields{
				awsRegion: EUCentral1,
				writer: redis.NewClient(&redis.Options{
					Addr: "writer_client_east",
				}),
				readers: mappedRedisRegions{
					USEast1: redis.NewClient(&redis.Options{
						Addr: "reader_client_east",
					}),
					USWest1: redis.NewClient(&redis.Options{
						Addr: "reader_client_west",
					}),
				},
			},
			args: args{operation: Read},
			want: redis.NewClient(&redis.Options{
				Addr: "writer_client_east",
			}),
		},
		{
			name: "should pick the writer even though the operation is set to READ because we are in the master location",
			fields: fields{
				awsRegion: USEast1,
				writer: redis.NewClient(&redis.Options{
					Addr: "writer_client_east",
				}),
				readers: mappedRedisRegions{
					USEast1: redis.NewClient(&redis.Options{
						Addr: "reader_client_east",
					}),
					USWest1: redis.NewClient(&redis.Options{
						Addr: "reader_client_west",
					}),
				},
			},
			args: args{operation: Read},
			want: redis.NewClient(&redis.Options{
				Addr: "writer_client_east",
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Configuration{
				localEnvironmentAWSRegion: tt.fields.awsRegion,
				writer:                    tt.fields.writer,
				readers:                   tt.fields.readers,
			}
			if got := c.RetrieveRedisClient(tt.args.operation); got.Options().Addr != tt.want.Options().Addr {
				t.Errorf("RetrieveRedisClient() = %v, want %v", got.Options().Addr, tt.want.Options().Addr)
			}
		})
	}
}

//BenchmarkName_RetrieveRedisClientWriter-12                 	1000000000	         0.511 ns/op	       0 B/op	       0 allocs/op
//BenchmarkName_RetrieveRedisClientWriter-12                 	1000000000	         0.510 ns/op	       0 B/op	       0 allocs/op
//BenchmarkName_RetrieveRedisClientWriter-12                 	1000000000	         0.509 ns/op	       0 B/op	       0 allocs/op
func BenchmarkName_RetrieveRedisClientWriter(b *testing.B) {
	var client *redis.Client
	_ = client
	c := &Configuration{
		localEnvironmentAWSRegion: USEast1,
		writer: redis.NewClient(&redis.Options{
			Addr: "writer_client",
		}),
		readers: mappedRedisRegions{
			USEast1: redis.NewClient(&redis.Options{
				Addr: "reader_client_east",
			}),
			USWest1: redis.NewClient(&redis.Options{
				Addr: "reader_client_west",
			}),
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client = c.RetrieveRedisClient(Write)
	}
}

//BenchmarkName_RetrieveRedisClientReader-12                 	1000000000	         0.328 ns/op	       0 B/op	       0 allocs/op
//BenchmarkName_RetrieveRedisClientReader-12                 	1000000000	         0.334 ns/op	       0 B/op	       0 allocs/op
//BenchmarkName_RetrieveRedisClientReader-12                 	1000000000	         0.330 ns/op	       0 B/op	       0 allocs/op
func BenchmarkName_RetrieveRedisClientReadWithWriter(b *testing.B) {
	var client *redis.Client
	_ = client
	c := &Configuration{
		localEnvironmentAWSRegion: USEast1,
		writer: redis.NewClient(&redis.Options{
			Addr: "writer_client",
		}),
		readers: mappedRedisRegions{
			USEast1: redis.NewClient(&redis.Options{
				Addr: "reader_client_east",
			}),
			USWest1: redis.NewClient(&redis.Options{
				Addr: "reader_client_west",
			}),
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client = c.RetrieveRedisClient(Read)
	}
}

//BenchmarkName_RetrieveRedisClientReader-12                 	308930834	         3.88 ns/op	       0 B/op	       0 allocs/op
//BenchmarkName_RetrieveRedisClientReader-12                 	309728967	         3.86 ns/op	       0 B/op	       0 allocs/op
//BenchmarkName_RetrieveRedisClientReader-12                 	310779890	         3.87 ns/op	       0 B/op	       0 allocs/op
func BenchmarkName_RetrieveRedisClientReader(b *testing.B) {
	var client *redis.Client
	_ = client
	c := &Configuration{
		localEnvironmentAWSRegion: USWest1,
		writer: redis.NewClient(&redis.Options{
			Addr: "writer_client",
		}),
		readers: mappedRedisRegions{
			USEast1: redis.NewClient(&redis.Options{
				Addr: "reader_client_east",
			}),
			USWest1: redis.NewClient(&redis.Options{
				Addr: "reader_client_west",
			}),
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client = c.RetrieveRedisClient(Read)
	}
}

//BenchmarkName_RetrieveRedisClientWriterDefaultReader-12    	196845808	         6.06 ns/op	       0 B/op	       0 allocs/op
//BenchmarkName_RetrieveRedisClientWriterDefaultReader-12    	196744748	         6.06 ns/op	       0 B/op	       0 allocs/op
//BenchmarkName_RetrieveRedisClientWriterDefaultReader-12    	196640210	         6.09 ns/op	       0 B/op	       0 allocs/op
func BenchmarkName_RetrieveRedisClientWriterDefaultReader(b *testing.B) {
	var client *redis.Client
	_ = client
	c := &Configuration{
		localEnvironmentAWSRegion: EUCentral1,
		writer: redis.NewClient(&redis.Options{
			Addr: "writer_client",
		}),
		readers: mappedRedisRegions{
			USEast1: redis.NewClient(&redis.Options{
				Addr: "reader_client_east",
			}),
			USWest1: redis.NewClient(&redis.Options{
				Addr: "reader_client_west",
			}),
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client = c.RetrieveRedisClient(Read)
	}
}
