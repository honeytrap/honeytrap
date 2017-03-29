package tests

import "testing"

func subTestSearch(t *testing.T, mc *mockServer) {
	runStep(t, mc, "KNN", keys_KNN_test)
}

func keys_KNN_test(mc *mockServer) error {
	return mc.DoBatch([][]interface{}{
		{"SET", "mykey", "1", "POINT", 5, 5}, {"OK"},
		{"SET", "mykey", "2", "POINT", 19, 19}, {"OK"},
		{"SET", "mykey", "3", "POINT", 12, 19}, {"OK"},
		{"SET", "mykey", "4", "POINT", -5, 5}, {"OK"},
		{"SET", "mykey", "5", "POINT", 33, 21}, {"OK"},
		{"NEARBY", "mykey", "LIMIT", 10, "DISTANCE", "POINTS", "POINT", 20, 20}, {
			"[0 [" +
				"[2 [19 19] 152808.67164037024] " +
				"[3 [12 19] 895945.1409106688] " +
				"[5 [33 21] 1448929.5916252395] " +
				"[1 [5 5] 2327116.1069888202] " +
				"[4 [-5 5] 3227402.6159841116]" +
				"]]"},
	})
}
