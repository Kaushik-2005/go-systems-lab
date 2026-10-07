package shard

import (
	"errors"
	"sort"
)

var ErrInvalidRanges = errors.New("ranges must be ordered, non-overlapping, and cover the key")

type KeyRange struct {
	Start string
	End   string
	Shard string
}

type RangeRouter struct {
	ranges []KeyRange
}

func NewRangeRouter(ranges []KeyRange) (*RangeRouter, error) {
	rangeCopy := append([]KeyRange(nil), ranges...)
	sort.Slice(rangeCopy, func(i, j int) bool {
		return rangeCopy[i].Start < rangeCopy[j].Start
	})

	for index, current := range rangeCopy {
		if current.Shard == "" || current.End != "" && current.Start >= current.End {
			return nil, ErrInvalidRanges
		}
		if index > 0 && rangeCopy[index-1].End != "" && current.Start < rangeCopy[index-1].End {
			return nil, ErrInvalidRanges
		}
	}
	return &RangeRouter{ranges: rangeCopy}, nil
}

func (r *RangeRouter) Route(key string) (string, error) {
	for _, current := range r.ranges {
		if key < current.Start {
			continue
		}
		if current.End == "" || key < current.End {
			return current.Shard, nil
		}
	}
	return "", ErrInvalidRanges
}

func (r *RangeRouter) Ranges() []KeyRange {
	return append([]KeyRange(nil), r.ranges...)
}
