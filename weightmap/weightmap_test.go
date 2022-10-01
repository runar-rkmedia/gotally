package weightmap

import "testing"

func Test_weightMap_Get(t *testing.T) {
	tests := []struct {
		name      string
		c         *weightMap
		loopCount int
		// key is the value that should be returned, with the values being min and max distribution-range
		distribution map[int][2]float64
	}{
		// TODO: Add test cases.
		{
			"Numbers should be distributed according to map (uneven)",
			NewWeightMap().
				Add(100, 42).
				Add(200, 17),
			1e4,
			map[int][2]float64{
				42: {0.3, 0.38},
				17: {0.63, 0.69},
			},
		},
		{
			"Numbers should be distributed according to map (even@100)",
			NewWeightMap().
				Add(100, 42).
				Add(100, 17).
				Add(100, 12).
				Add(100, 6),
			1e3,
			map[int][2]float64{
				42: {0.2, 0.3},
				17: {0.2, 0.3},
				12: {0.2, 0.3},
				6:  {0.2, 0.3},
			},
		},
		{
			"Numbers should be distributed according to map (even@1)",
			NewWeightMap().
				Add(1, 42).
				Add(1, 17).
				Add(1, 6).
				Add(1, 12),
			1e3,
			map[int][2]float64{
				42: {0.2, 0.3},
				17: {0.2, 0.3},
				12: {0.2, 0.3},
				6:  {0.2, 0.3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distributionMap := map[int]int{}
			for i := 0; i < tt.loopCount; i++ {
				got := tt.c.Get(i)
				distributionMap[got]++
			}
			distribution := map[int]float64{}
			for k, v := range distributionMap {
				distribution[k] = float64(v) / float64(tt.loopCount)
			}
			t.Log(distribution)
			for k, v := range tt.distribution {
				min := v[0]
				max := v[1]
				if distribution[k] < min {
					t.Errorf("distribution is lower than expected for\t%02d: want %.2f-%.2f, got %.4f", k, min, max, distribution[k])
				}
				if distribution[k] > max {
					t.Errorf("distribution is higher than expected for\t%02d: want %.2f-%.2f, got %.4f", k, min, max, distribution[k])
				}

			}
		})
	}
}
