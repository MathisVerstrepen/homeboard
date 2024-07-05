package classification

import (
	"bufio"
	"fmt"
	"image"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/disintegration/imaging"
	tf "github.com/galeone/tensorflow/tensorflow/go"
	tg "github.com/galeone/tfgo"
)

var (
	model  *tg.Model
	labels []string
)

type Classification struct {
	Label      string
	Proability float32
}

func loadLabels(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func convertValue(value uint32) float32 {
	return (float32(value >> 8)) / float32(255)
}

func imageToTensor(img image.Image, imageHeight, imageWidth int) (tfTensor *tf.Tensor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("classify: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if imageHeight <= 0 || imageWidth <= 0 {
		return tfTensor, fmt.Errorf("classify: image width and height must be > 0")
	}

	var tfImage [1][][][3]float32

	for j := 0; j < imageHeight; j++ {
		tfImage[0] = append(tfImage[0], make([][3]float32, imageWidth))
	}

	for i := 0; i < imageWidth; i++ {
		for j := 0; j < imageHeight; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			tfImage[0][j][i][0] = convertValue(r)
			tfImage[0][j][i][1] = convertValue(g)
			tfImage[0][j][i][2] = convertValue(b)
		}
	}
	return tf.NewTensor(tfImage)
}

func createTensor(imagePath string) (*tf.Tensor, error) {
	fmt.Println("/home/mverstre/Documents/Dev/homeboard/" + imagePath)
	srcImage, err := imaging.Open("/home/mverstre/Documents/Dev/homeboard/" + imagePath)
	if err != nil {
		return nil, err
	}
	img := imaging.Fill(srcImage, 224, 224, imaging.Center, imaging.Lanczos)
	return imageToTensor(img, 224, 224)
}

func LabelizeImageNSFW(imagePath string) ([]Classification, error) {
	//Gets rid of some annoying tensorflow warnings
	os.Setenv("TF_CPP_MIN_LOG_LEVEL", "2")

	modelName := "/home/mverstre/Documents/Dev/homeboard/assets/models/mobilenet_v2_140_224"
	loadLabels(modelName + "/class_labels.txt")
	model = tg.LoadModel(modelName, []string{"serve"}, nil)

	normalizedImg, err := createTensor(imagePath)
	if err != nil {
		log.Printf("unable to make a normalizedImg from image: %v", err)
		return nil, err
	}

	results := model.Exec(
		[]tf.Output{
			model.Op("StatefulPartitionedCall", 0),
		},
		map[tf.Output]*tf.Tensor{
			model.Op("serving_default_input", 0): normalizedImg,
		},
	)

	probabilities := results[0].Value().([][]float32)[0]
	classifications := []Classification{}

	if len(probabilities) == 0 {
		fmt.Println("no probabilities returned from model")
		return nil, fmt.Errorf("no probabilities returned from model")
	}
	if len(labels) == 0 {
		fmt.Println("no labels loaded")
		return nil, fmt.Errorf("no labels loaded")
	}

	for i, p := range probabilities {
		// if p < 5 {
		// 	continue
		// }
		classifications = append(classifications, Classification{
			Label:      strings.ToLower(labels[i]),
			Proability: p,
		})
	}

	return classifications, nil
}

func LabelizeImage(imagePath string) ([]Classification, error) {
	//Gets rid of some annoying tensorflow warnings
	os.Setenv("TF_CPP_MIN_LOG_LEVEL", "2")

	modelName := "/home/mverstre/Documents/Dev/homeboard/assets/models/mobilenet-v2-tensorflow2-100-224-classification-v2"
	loadLabels(modelName + "/ImageNetLabels.txt")
	model = tg.LoadModel(modelName, []string{"serve"}, nil)

	normalizedImg, err := createTensor(imagePath)
	if err != nil {
		log.Printf("unable to make a normalizedImg from image: %v", err)
		return nil, err
	}

	results := model.Exec(
		[]tf.Output{
			model.Op("StatefulPartitionedCall", 0),
		},
		map[tf.Output]*tf.Tensor{
			model.Op("serving_default_inputs", 0): normalizedImg,
		},
	)

	probabilities := results[0].Value().([][]float32)[0]
	classifications := []Classification{}

	if len(probabilities) == 0 {
		fmt.Println("no probabilities returned from model")
		return nil, fmt.Errorf("no probabilities returned from model")
	}
	if len(labels) == 0 {
		fmt.Println("no labels loaded")
		return nil, fmt.Errorf("no labels loaded")
	}

	for i, p := range probabilities {
		if p < 5 {
			continue
		}
		classifications = append(classifications, Classification{
			Label:      strings.ToLower(labels[i]),
			Proability: p,
		})
	}

	return classifications, nil
}

func IsImageNsfw(results []Classification) bool {
	unsafeCount := 0

	for _, classification := range results {
		if classification.Label != "neutral" && classification.Label != "drawings" {
			if classification.Proability > 0.3 {
				unsafeCount += 1
			}
		}
	}

	return unsafeCount > 0
}
