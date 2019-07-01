package strobe

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// Operation is holding a test vector operation
type Operation struct {
	OpName string `json:"name"`

	// for Init
	OpCustomString string `json:"custom_string,omitempty"`
	OpSecurity     int    `json:"security,omitempty"`

	// for other operations
	OpMeta        bool   `json:"meta"`
	OpInputData   string `json:"input_data,omitempty"`
	OpInputLength int    `json:"input_length,omitempty"`
	OpOutput      string `json:"output,omitempty"`
	OpStateAfter  string `json:"state_after"`
	OpStream      bool   `json:"stream"`
}

func DebugInit(customString string, security int) (_ Strobe, op Operation) {

	s := InitStrobe(customString, security)

	op.OpName = "init"
	op.OpCustomString = customString
	op.OpSecurity = 128
	op.OpStateAfter = s.debugPrintState()

	//
	return s, op
}

func (s *Strobe) DebugGoThroughOperation(operation string, meta bool, inputData []byte, inputLength int, stream bool) (op Operation) {
	// create operation object
	op.OpName = operation
	op.OpInputData = hex.EncodeToString(inputData)
	op.OpInputLength = inputLength
	op.OpMeta = meta
	op.OpStream = stream
	// go through operation
	outputData := s.Operate(meta, operation, inputData, inputLength, stream)
	if len(outputData) > 0 {
		op.OpOutput = hex.EncodeToString(outputData)
	}
	// state
	op.OpStateAfter = s.debugPrintState()
	//
	return
}

type TestVector struct {
	Name       string      `json:"name"`
	Operations []Operation `json:"operations"`
}

type TestVectors struct {
	TestVectors []TestVector `json:"test_vectors"`
}

func TestGenTestVectors(t *testing.T) {
	// skipping this
	if testing.Short() {
		t.Skip("skipping generation of test vectors.")
	}

	// test vector file
	out, err := os.Create("test_vectors/test_vectors.json")
	if err != nil {
		t.Fatal("couldn't create test vector file")
	}
	defer out.Close()

	// the structure
	var testVectors TestVectors

	// go through runs
	testVectors.TestVectors = append(testVectors.TestVectors, simpleTest())
	testVectors.TestVectors = append(testVectors.TestVectors, metaTest())
	testVectors.TestVectors = append(testVectors.TestVectors, streamingTest())
	testVectors.TestVectors = append(testVectors.TestVectors, boundaryBlocksTest())

	// save output
	jsonOutput, _ := json.Marshal(testVectors)
	fmt.Fprintf(out, string(jsonOutput))
}

func simpleTest() (testVector TestVector) {
	// start the run
	testVector.Name = "simple tests"
	// init
	s, op := DebugInit("custom string", 128)
	testVector.Operations = append(testVector.Operations, op)
	// KEY
	key := []byte("010101")
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("KEY", false, key, 0, false))
	// AD
	message := []byte("hello, how are you good sir?")
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("AD", false, message, 0, false))
	// PRF
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("PRF", false, []byte{}, 16, false))
	// send_ENC
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("send_ENC", false, []byte("hi how are you"), 0, false))
	// recv_ENC
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("recv_ENC", false, []byte("hi how are you"), 0, false))
	// send_MAC
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("send_MAC", false, []byte{}, 16, false))
	// recv_MAC
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("recv_MAC", false, []byte("hi how are you"), 0, false))
	// send_CLR
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("send_CLR", false, []byte("hi how are you"), 0, false))
	// recv_CLR
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("recv_CLR", false, []byte("hi how are you"), 0, false))
	// RATCHET
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("RATCHET", false, []byte{}, 32, false))
	return
}
func metaTest() (testVector TestVector) {
	// start the run
	testVector.Name = "meta tests"
	// init
	s, op := DebugInit("custom string number 2, that's a pretty long string", 128)
	testVector.Operations = append(testVector.Operations, op)
	// KEY
	key := []byte("010101")
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("KEY", false, key, 0, false))
	// AD
	message := []byte("hello, how are you good sir?")
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("AD", true, message, 0, false))
	// PRF
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("PRF", false, []byte{}, 16, false))
	// send_ENC
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("send_ENC", true, []byte("hi how are you"), 0, false))
	// recv_ENC
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("recv_ENC", true, []byte("hi how are you"), 0, false))
	// send_MAC
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("send_MAC", true, []byte{}, 16, false))
	// recv_MAC
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("recv_MAC", true, []byte("hi how are you"), 0, false))
	// send_CLR
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("send_CLR", true, []byte("hi how are you"), 0, false))
	// recv_CLR
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("recv_CLR", true, []byte("hi how are you"), 0, false))
	// RATCHET
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("RATCHET", true, []byte{}, 32, false))
	return
}

func streamingTest() (testVector TestVector) {
	// start the run
	testVector.Name = "streaming tests"
	// init
	s, op := DebugInit("custom string number 2, that's a pretty long string", 128)
	testVector.Operations = append(testVector.Operations, op)
	// KEY
	key := []byte("0101010100100101010101010101001001")
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("KEY", false, key, 0, false))
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("KEY", false, key, 0, true))
	// AD
	message := []byte("hello, how are you good sir? ????")
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("AD", false, message, 0, false))
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("AD", false, message, 0, true))
	testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("AD", false, message, 0, false))
	return
}

func boundaryBlocksTest() (testVector TestVector) {
	// start the run
	testVector.Name = "boundary tests"
	// init
	s, op := DebugInit("custom string number 2, that's a pretty long string", 128)
	testVector.Operations = append(testVector.Operations, op)
	// StrobeR is 166 for 128-bit security, so we go around it
	inputData := []byte{}
	for i := 0; i < 168; i++ { //
		inputData = append(inputData, byte(i))
		if i%3 == 0 {
			testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("KEY", false, inputData, 0, false))
		} else if i%3 == 1 {
			testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("AD", false, inputData, 0, false))
		} else {
			testVector.Operations = append(testVector.Operations, s.DebugGoThroughOperation("send_ENC", false, inputData, 0, false))
		}
	}
	return
}
