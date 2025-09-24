package memory

import (
	"testing"
)

func TestParseStreamOutput(t *testing.T) {
	// Sample STREAM output based on the problem statement
	sampleOutput := `-------------------------------------------------------------
STREAM version $Revision: 5.10 $
-------------------------------------------------------------
This system uses 8 bytes per array element.
-------------------------------------------------------------
Array size = 10000000 (elements), Offset = 0 (elements)
Memory per array = 76.3 MiB (= 0.1 GiB).
Total memory required = 228.9 MiB (= 0.2 GiB).
Each kernel will be executed 10 times.
 The *best* time for each kernel (excluding the first iteration)
 will be used to compute the reported bandwidth.
-------------------------------------------------------------
Your clock granularity/precision appears to be 1 microseconds.
Each test below will take on the order of 1008115 microseconds.
   (= 1008115 clock ticks)
Increase the size of the arrays if this shows that
you are not getting at least 20 clock ticks per test.
-------------------------------------------------------------
WARNING -- The above is only a rough guideline.
For best results, please be sure you know the
precision of your system timer.
-------------------------------------------------------------
Function    Best Rate MB/s  Avg time     Min time     Max time
Copy:           21792.8     0.011733     0.007342     0.031549
Scale:          14821.8     0.019170     0.010795     0.051002
Add:            16917.9     0.026095     0.014186     0.058414
Triad:          17097.8     0.024922     0.014037     0.049033
-------------------------------------------------------------
Solution Validates: avg error less than 1.000000e-13 on all three arrays
-------------------------------------------------------------`

	expected := `Function    Best Rate MB/s  Avg time     Min time     Max time
Copy:           21792.8     0.011733     0.007342     0.031549
Scale:          14821.8     0.019170     0.010795     0.051002
Add:            16917.9     0.026095     0.014186     0.058414
Triad:          17097.8     0.024922     0.014037     0.049033
`

	result := parseStreamOutput(sampleOutput, "en")
	
	if result != expected {
		t.Errorf("parseStreamOutput failed.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestParseStreamOutputNoFunction(t *testing.T) {
	// Test with output that doesn't contain Function section
	sampleOutput := `-------------------------------------------------------------
STREAM version $Revision: 5.10 $
-------------------------------------------------------------
Some other output without Function section
-------------------------------------------------------------`

	result := parseStreamOutput(sampleOutput, "en")
	
	if result != "" {
		t.Errorf("parseStreamOutput should return empty string for invalid output, got: %s", result)
	}
}