package util

import (
	"fmt"
	"reflect"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/azer/snakecase"
	"github.com/krallistic/kafka-operator/spec"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
)

var ()

func ReflectOptionsStruct(v interface{}) []v1.EnvVar {
	val := reflect.ValueOf(v)
	options := make([]v1.EnvVar, 0)

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		if valueField.Kind() == reflect.Interface && !valueField.IsNil() {
			elm := valueField.Elem()
			if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
				valueField = elm
			}
		}

		if valueField.IsNil() {
			continue
		} else {
			if valueField.Kind() == reflect.Ptr {
				valueField = valueField.Elem()
			}
			value := fmt.Sprintf("%v", valueField.Interface())
			name := strings.ToUpper(snakecase.SnakeCase(typeField.Name))

			env := v1.EnvVar{
				Name:  name,
				Value: value,
			}
			options = append(options, env)
		}
	}
	return options
}

func (c *ClientUtil) GetMaxHeap(cluster spec.Kafkacluster) *resource.Quantity {
	memory, err := resource.ParseQuantity(cluster.Spec.Resources.Memory)
	if err != nil {
		memory, _ = resource.ParseQuantity(defaultMemory)
	}

	heapsize := int64(float64(memory.ScaledValue(resource.Mega)) * 0.6)
	if heapsize > 4096 {
		heapsize = 4096
	}
	maxHeap := resource.NewScaledQuantity(heapsize, resource.Mega)

	return maxHeap
}

func (c *ClientUtil) GetMaxHeapJavaString(cluster spec.Kafkacluster) string {
	return "-Xmx" + c.GetMaxHeap(cluster).String()
}

func (c *ClientUtil) GenerateKafkaOptions(cluster spec.Kafkacluster) []v1.EnvVar {
	kafkaOptions := cluster.Spec.KafkaOptions

	structOptions := ReflectOptionsStruct(kafkaOptions)
	logger.WithFields(log.Fields{
		"method":  "GenerateKafkaOptions",
		"options": structOptions,
	}).Debug("Generated KafkaOptions from Struct to Env Vars")

	staticOptions := []v1.EnvVar{
		v1.EnvVar{
			Name: "NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		v1.EnvVar{
			Name:  "KAFKA_ZOOKEEPER_CONNECT",
			Value: cluster.Spec.ZookeeperConnect,
		},
		v1.EnvVar{
			Name:  "KAFKA_HEAP_OPTS",
			Value: c.GetMaxHeapJavaString(cluster),
		},
	}

	options := append(structOptions, staticOptions...)
	fmt.Println(options)
	return options
}
