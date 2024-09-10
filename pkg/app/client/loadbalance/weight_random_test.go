/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package loadbalance

import (
	"math"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

func TestWeightedBalancer(t *testing.T) {
	balancer := NewWeightedBalancer()
	// nil
	ins := balancer.Pick(discovery.Result{})
	assert.DeepEqual(t, ins, nil)

	// empty instance
	e := discovery.Result{
		Instances: make([]discovery.Instance, 0),
		CacheKey:  "a",
	}
	balancer.Rebalance(e)
	ins = balancer.Pick(e)
	assert.DeepEqual(t, ins, nil)

	// one instance
	insList := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8888", 20, nil),
	}
	e = discovery.Result{
		Instances: insList,
		CacheKey:  "b",
	}
	balancer.Rebalance(e)
	for i := 0; i < 100; i++ {
		ins = balancer.Pick(e)
		assert.DeepEqual(t, ins.Weight(), 20)
	}

	// multi instances, weightSum > 0
	insList = []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8881", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8882", 20, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8883", 50, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8884", 100, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8885", 200, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8886", 500, nil),
	}

	var weightSum int
	for _, ins := range insList {
		weight := ins.Weight()
		weightSum += weight
	}

	n := 10000000
	pickedStat := map[int]int{}
	e = discovery.Result{
		Instances: insList,
		CacheKey:  "c",
	}
	balancer.Rebalance(e)
	for i := 0; i < n; i++ {
		ins = balancer.Pick(e)
		weight := ins.Weight()
		if pickedCnt, ok := pickedStat[weight]; ok {
			pickedStat[weight] = pickedCnt + 1
		} else {
			pickedStat[weight] = 1
		}
	}

	for _, ins := range insList {
		weight := ins.Weight()
		expect := float64(weight) / float64(weightSum) * float64(n)
		actual := float64(pickedStat[weight])
		delta := math.Abs(expect - actual)
		assert.DeepEqual(t, true, delta/expect < 0.01)
	}

	// have instances that weight < 0
	insList = []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8881", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8882", -10, nil),
	}
	e = discovery.Result{
		Instances: insList,
		CacheKey:  "d",
	}
	balancer.Rebalance(e)
	for i := 0; i < 1000; i++ {
		ins = balancer.Pick(e)
		assert.DeepEqual(t, 10, ins.Weight())
	}
}

func TestWeightedBalancerDeleteAndName(t *testing.T) {
	balancer := NewWeightedBalancer()

	// Test Name method
	assert.DeepEqual(t, "weight_random", balancer.Name())

	// Test Delete method
	insList := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8881", 10, nil),
	}
	e := discovery.Result{
		Instances: insList,
		CacheKey:  "test_delete",
	}
	balancer.Rebalance(e)

	// Verify instance is picked before deletion
	ins := balancer.Pick(e)
	assert.NotNil(t, ins)
	assert.DeepEqual(t, "127.0.0.1:8881", ins.Address().String())

	// Delete the cache entry
	balancer.Delete("test_delete")

	// Verify no instance is picked after deletion
	ins = balancer.Pick(e)
	assert.Nil(t, ins)
}

func TestWeightedBalancerRebalanceAndPickEdgeCases(t *testing.T) {
	balancer := NewWeightedBalancer()

	// Test Rebalance with empty instance list
	emptyResult := discovery.Result{
		Instances: []discovery.Instance{},
		CacheKey:  "empty",
	}
	balancer.Rebalance(emptyResult)
	ins := balancer.Pick(emptyResult)
	assert.Nil(t, ins)

	// Test Rebalance with instances having zero or negative weights
	mixedWeightInsList := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8881", 0, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8882", -5, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8883", 10, nil),
	}
	mixedWeightResult := discovery.Result{
		Instances: mixedWeightInsList,
		CacheKey:  "mixed_weight",
	}
	balancer.Rebalance(mixedWeightResult)

	// Verify that only instances with positive weights are picked
	pickedAddresses := make(map[string]int)
	totalPicks := 1000
	for i := 0; i < totalPicks; i++ {
		ins = balancer.Pick(mixedWeightResult)
		assert.NotNil(t, ins)
		pickedAddresses[ins.Address().String()]++
	}
	assert.DeepEqual(t, 1, len(pickedAddresses))
	assert.DeepEqual(t, totalPicks, pickedAddresses["127.0.0.1:8883"])
	assert.DeepEqual(t, 0, pickedAddresses["127.0.0.1:8881"])
	assert.DeepEqual(t, 0, pickedAddresses["127.0.0.1:8882"])

	// Test Pick with non-existent cache key
	nonExistentResult := discovery.Result{
		Instances: []discovery.Instance{
			discovery.NewInstance("tcp", "127.0.0.1:8884", 10, nil),
		},
		CacheKey: "non_existent",
	}
	ins = balancer.Pick(nonExistentResult)
	assert.Nil(t, ins)

	// Test Rebalance with all zero or negative weights
	allZeroOrNegativeInsList := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8885", 0, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8886", -5, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8887", -10, nil),
	}
	allZeroOrNegativeResult := discovery.Result{
		Instances: allZeroOrNegativeInsList,
		CacheKey:  "all_zero_or_negative",
	}
	balancer.Rebalance(allZeroOrNegativeResult)
	ins = balancer.Pick(allZeroOrNegativeResult)
	assert.Nil(t, ins)

	// Test Pick after Rebalance
	balancer.Rebalance(mixedWeightResult)
	ins = balancer.Pick(mixedWeightResult)
	assert.NotNil(t, ins)
	assert.DeepEqual(t, "127.0.0.1:8883", ins.Address().String())

	// Test Delete
	balancer.Delete(mixedWeightResult.CacheKey)
	ins = balancer.Pick(mixedWeightResult)
	assert.Nil(t, ins)

	// Test Rebalance and Pick with mixed weights including positive ones
	mixedWeightsWithPositive := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8881", 0, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8882", -5, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8883", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8884", 20, nil),
	}
	mixedWeightsWithPositiveResult := discovery.Result{
		Instances: mixedWeightsWithPositive,
		CacheKey:  "mixed_weights_with_positive",
	}
	balancer.Rebalance(mixedWeightsWithPositiveResult)

	pickedAddresses = make(map[string]int)
	totalPicks = 1000000 // Increased for better statistical accuracy
	for i := 0; i < totalPicks; i++ {
		ins = balancer.Pick(mixedWeightsWithPositiveResult)
		assert.NotNil(t, ins)
		pickedAddresses[ins.Address().String()]++
	}

	// Verify that only instances with positive weights are picked
	assert.DeepEqual(t, 2, len(pickedAddresses))
	assert.DeepEqual(t, 0, pickedAddresses["127.0.0.1:8881"])
	assert.DeepEqual(t, 0, pickedAddresses["127.0.0.1:8882"])
	assert.DeepEqual(t, true, pickedAddresses["127.0.0.1:8883"] > 0)
	assert.DeepEqual(t, true, pickedAddresses["127.0.0.1:8884"] > 0)
	assert.DeepEqual(t, totalPicks, pickedAddresses["127.0.0.1:8883"]+pickedAddresses["127.0.0.1:8884"])

	// Verify the distribution of picked addresses
	totalWeight := 30 // 10 + 20
	expectedRatio8883 := float64(10) / float64(totalWeight)
	expectedRatio8884 := float64(20) / float64(totalWeight)
	actualRatio8883 := float64(pickedAddresses["127.0.0.1:8883"]) / float64(totalPicks)
	actualRatio8884 := float64(pickedAddresses["127.0.0.1:8884"]) / float64(totalPicks)

	// Allow for a 2% margin of error in the distribution
	marginOfError := 0.02
	assert.DeepEqual(t, true, math.Abs(expectedRatio8883-actualRatio8883) < marginOfError)
	assert.DeepEqual(t, true, math.Abs(expectedRatio8884-actualRatio8884) < marginOfError)
}
