package models

type HashCacheConfig struct {
	HashCacheEntryExpireUnixTime int64 `mapstructure:"hash_cache_entry_expire_unix_time"`
}

func (c *HashCacheConfig) Validate() error {
	if c.HashCacheEntryExpireUnixTime <= 0 {
		return ErrInvalidHashCacheEntryExpireUnixTime
	}

	return nil
}
