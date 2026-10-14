package server

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

const minWorldIndex = 0
const maxWorldIndex = 9

var (
	ErrInvalidIndex = status.Errorf(codes.InvalidArgument, "world index must be between %d and %d", minWorldIndex, maxWorldIndex)
	ErrWorldExists  = status.Errorf(codes.AlreadyExists, "world at that index already exists")
)

func getNextWorldIndex() (int, error) {
	entries, err := os.ReadDir("/tmp/homelab/k8s/minecraft/worlds/")
	if err != nil {
		return -1, status.Errorf(codes.Internal, "failed to read world files: %v", err)
	}

	taken := make(map[int]bool)
	for _, entry := range entries {
		name := strings.TrimPrefix(strings.TrimSuffix(entry.Name(), ".yaml"), "minecraft-")
		index, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		taken[index] = true
	}

	for i := minWorldIndex; i <= maxWorldIndex; i++ {
		if !taken[i] {
			return i, nil
		}
	}

	return -1, status.Errorf(codes.ResourceExhausted, "no available world slots")
}

type MinecraftWorldConfig struct {
	Name      string `yaml:"name"`
	ClusterIP string `yaml:"clusterIP"`
}

func createMinecraftYAML(index int) error {
	if index < minWorldIndex || index > maxWorldIndex {
		return ErrInvalidIndex
	}

	worldName := fmt.Sprintf("minecraft-%d", index)
	filePath := fmt.Sprintf("/tmp/homelab/k8s/minecraft/worlds/%s.yaml", worldName)

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrWorldExists
		}
		return status.Errorf(codes.Internal, "failed to create file: %v", err)
	}
	defer file.Close()

	config := MinecraftWorldConfig{
		Name:      worldName,
		ClusterIP: fmt.Sprintf("10.96.1.%d", index),
	}

	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to marshal yaml: %v", err)
	}

	if _, err := file.Write(yamlData); err != nil {
		return status.Errorf(codes.Internal, "failed to write file: %v", err)
	}

	return nil
}
