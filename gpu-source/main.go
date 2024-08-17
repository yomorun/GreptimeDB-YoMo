package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/yomorun/yomo"
)

var (
	host       = "local"
	tag        = uint32(0xC010)
	credential = ""
	zipperAddr = "127.0.0.1:9000"
	source     yomo.Source
)

func init() {
	host, _ = os.Hostname()

	if v, ok := os.LookupEnv("ZIPPER_ADDR"); ok {
		zipperAddr = v
	}
	log.Println("ZIPPER_ADDR: ", zipperAddr)

	if v, ok := os.LookupEnv("CREDENTIAL"); ok {
		credential = v
	}
	log.Println("CREDENTIAL: ", credential)

	ss, err := newSource(zipperAddr, credential)
	if err != nil {
		log.Fatalln(err)
	}
	source = ss
}

func main() {
	err := InitNvml()
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		source.Close()

		err = ShutdownNvml()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	for {
		if err := Run(); err != nil {
			log.Fatalln(err)
		}
		time.Sleep(5 * time.Second)
	}

}

func newSource(zipperAddr string, credential string) (yomo.Source, error) {
	opts := []yomo.SourceOption{
		yomo.WithSourceReConnect(),
	}
	if credential != "" {
		opts = append(opts, yomo.WithCredential(credential))
	}
	source := yomo.NewSource(
		"gpu-collector",
		zipperAddr,
		opts...,
	)
	if err := source.Connect(); err != nil {
		return nil, source.Connect()
	}
	return source, nil
}

func InitNvml() error {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return errors.New(nvml.ErrorString(ret))
	}
	return nil
}

func ShutdownNvml() error {
	ret := nvml.Shutdown()
	if ret != nvml.SUCCESS {
		return errors.New(nvml.ErrorString(ret))
	}
	return nil
}

func Run() error {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return errors.New(nvml.ErrorString(ret))
	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return errors.New(nvml.ErrorString(ret))
		}
		name, ret := device.GetName()
		if ret != nvml.SUCCESS {
			return errors.New(nvml.ErrorString(ret))
		}
		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			return errors.New(nvml.ErrorString(ret))
		}

		temperature, ret := nvml.DeviceGetTemperature(device, nvml.TEMPERATURE_GPU)
		if ret != nvml.SUCCESS {
			return errors.New(nvml.ErrorString(ret))
		}

		memory, ret := nvml.DeviceGetMemoryInfo_v2(device)
		if ret != nvml.SUCCESS {
			return errors.New(nvml.ErrorString(ret))
		}

		utilization, ret := nvml.DeviceGetUtilizationRates(device)
		if ret != nvml.SUCCESS {
			return errors.New(nvml.ErrorString(ret))
		}

		powerLimit, ret := nvml.DeviceGetPowerManagementLimit(device)
		if ret != nvml.SUCCESS {
			return errors.New(nvml.ErrorString(ret))
		}

		powerUsage, ret := nvml.DeviceGetPowerUsage(device)
		if ret != nvml.SUCCESS {
			return errors.New(nvml.ErrorString(ret))
		}

		line := fmt.Sprintf(
			"nvidia_gpu,host=%s,name=%s,uuid=%s temperature=%d,memory_total=%d,memory_used=%d,memory_free=%d,memory_reserved=%d,utilization_memory=%d,utilization_gpu=%d,power_limit=%d,power_usage=%d %d",
			host,
			strings.ReplaceAll(name, " ", "-"),
			uuid,
			temperature,
			memory.Total,
			memory.Used,
			memory.Free,
			memory.Reserved,
			utilization.Memory,
			utilization.Gpu,
			powerLimit,
			powerUsage,
			time.Now().UnixNano(),
		)

		fmt.Println(line)
		if err := source.Write(tag, []byte(line)); err != nil {
			log.Println("source write error:", err)
		}
	}

	return nil
}
