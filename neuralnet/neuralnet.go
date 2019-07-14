package neuralnet

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"gopkg.in/cheggaaa/pb.v1"
	"math/rand"
	"os"
	"time"
)

type NeuralNetwork struct {
	HiddenLayer      []float64
	InputLayer       []float64
	OutputLayer      []float64
	WeightHidden     [][]float64
	WeightOutput     [][]float64
	ErrOutput        []float64
	ErrHidden        []float64
	LastChangeHidden [][]float64
	LastChangeOutput [][]float64
	Rate1            float64
	Rate2            float64
}

const errorMessage = "The amount of input variable doesn't match."

func DumpNN(fileName string, nn *NeuralNetwork) {
	out_f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		panic("Failed to dump the network to " + fileName)
	}
	fmt.Println("hey")
	defer out_f.Close()
	encoder := json.NewEncoder(out_f)
	fmt.Println("hey2")
	err = encoder.Encode(nn)
	fmt.Println(nn)
	if err != nil {
		panic(err)
	}
}

func LoadNN(fileName string) *NeuralNetwork {
	in_f, err := os.Open(fileName)
	if err != nil {
		panic("Failed to load " + fileName)
	}
	defer in_f.Close()
	decoder := json.NewDecoder(in_f)
	nn := &NeuralNetwork{}
	err = decoder.Decode(nn)
	if err != nil {
		panic(err)
	}
	//fmt.Println(nn)
	return nn
}

// Returns the value of a neural network where rate1 is equal to 0.25 and rate2 to 0.1
func DefaultNetwork(input, hidden, output int) *NeuralNetwork {
	return NewNetwork(input, hidden, output, 0.25, 0.1)
}

func NewNetwork(input, hidden, output int, rate1, rate2 float64) *NeuralNetwork {
	rand.Seed(time.Now().UnixNano())

	input++
	hidden++

	return &NeuralNetwork{
		Rate1:            rate1,
		Rate2:            rate2,
		InputLayer:       make([]float64, input),
		HiddenLayer:      make([]float64, hidden),
		OutputLayer:      make([]float64, output),
		ErrOutput:        make([]float64, output),
		ErrHidden:        make([]float64, hidden),
		WeightHidden:     RandomMatrix(hidden, input, -1.0, 1.0),
		WeightOutput:     RandomMatrix(output, hidden, -1.0, 1.0),
		LastChangeHidden: MakeMatrix(hidden, input, 0.0),
		LastChangeOutput: MakeMatrix(output, hidden, 0.0),
	}
}

func (network *NeuralNetwork) Forward(input []float64) (output []float64) {
	if len(input) + 1 != len(network.InputLayer) {
		panic("The length of input values must match the number of input layers.")
	}

	// Set the first layer
	for i := 0; i < len(input); i++ {
		network.InputLayer[i] = input[i]
	}
	network.InputLayer[len(network.InputLayer) - 1] = 1.0 // Bias node for input layer

	// Apply the weights to the input layer to give the hidden layer
	hiddenLayer := ApplyWeights(
		len(network.WeightHidden) - 1,
		network.InputLayer,
		network.WeightHidden,
	)
	network.HiddenLayer = ApplyFunc(hiddenLayer, Sigmoid)

	network.HiddenLayer[len(network.HiddenLayer) - 1] = 1.0 // Bias node for hidden layer
	// Apply weights to the hidden layer to give the output layer
	network.OutputLayer = ApplyWeights(
		len(network.WeightOutput),
		network.HiddenLayer,
		network.WeightOutput,
	)

	return network.OutputLayer[:]
}

func (network *NeuralNetwork) Feedback(target []float64) {
	// Set the output errors
	for i, outputValue := range network.OutputLayer {
		network.ErrOutput[i] = outputValue - target[i]
	}

	// Calculate the hidden errors
	for i := 0; i < len(network.HiddenLayer)-1; i++ {
		err := 0.0
		for j := 0; j < len(network.OutputLayer); j++ {
			err += network.ErrOutput[j] * network.WeightOutput[j][i]
		}
		network.ErrHidden[i] = err
	}

	// Adjust output weights
	for i := 0; i < len(network.OutputLayer); i++ {
		for j := 0; j < len(network.HiddenLayer); j++ {
			delta := network.ErrOutput[i]
			change := network.Rate1 * delta * network.HiddenLayer[j] + network.Rate2 * network.LastChangeOutput[i][j]

			network.WeightOutput[i][j] -= change
			network.LastChangeOutput[i][j] = change
		}
	}

	// Adjust hidden weights
	for i := 0; i < len(network.HiddenLayer)-1; i++ {
		for j := 0; j < len(network.InputLayer); j++ {
			delta := network.ErrHidden[i] * SigmoidDerivative(network.HiddenLayer[i])
			change := network.Rate1 * delta * network.InputLayer[j] + network.Rate2 * network.LastChangeHidden[i][j]

			network.WeightHidden[i][j] -= change
			network.LastChangeHidden[i][j] = change
		}
	}
}

func (network *NeuralNetwork) CalculateError(target []float64) (errSum float64) {
	for i := 0; i < len(network.OutputLayer); i++ {
		err := network.OutputLayer[i] - target[i]
		errSum += 0.5 * err * err
	}

	return errSum
}

func RandomIdx(n int) []int {
	A := make([]int, n)
	for i := 0; i < n; i++ {
		A[i] = i
	}

	for i := 0; i < n; i++ {
		j := i + int(rand.Float64()*float64(n-i))
		A[i], A[j] = A[j], A[i]
	}
	return A
}

func (network *NeuralNetwork) Train(inputs [][]float64, targets [][]float64, iteration int) {
	if len(inputs[0]) + 1 != len(network.InputLayer) {
		panic(errorMessage)
	}
	if len(targets[0]) != len(network.OutputLayer) {
		panic(errorMessage)
	}

	blue := color.FgBlue.Render

	count := 100
	bar := pb.New(count).Postfix(fmt.Sprintf(" - %s", blue("Creating the neural network")))
	bar.Format("(██ )")
	bar.SetWidth(60)
	bar.ShowCounters = false
	bar.Start()

	iterFlag := -1
	for i := 0; i < iteration; i++ {
		idxAry := RandomIdx(len(inputs))
		curErr := 0.0
		for j := 0; j < len(inputs); j++ {
			network.Forward(inputs[idxAry[j]])
			network.Feedback(targets[idxAry[j]])
			curErr += network.CalculateError(targets[idxAry[j]])
			if (j+1)%1000 == 0 {
				if iterFlag != i {
					fmt.Println("")
					iterFlag = i
				}
			}
		}

		if (iteration >= count && (i+1)%(iteration/count) == 0) || iteration < count {
			bar.Increment()
		}
	}

	bar.Finish()
}