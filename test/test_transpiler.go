// Package main provides a simple test for bitwise transpilation analysis
package main

import (
	"fmt"
	"log"

	"github.com/faelmori/gastype/internal/transpiler"
)

func main() {
	fmt.Println("🔍 Testing GASType Bitwise Analysis...")

	// Create transpiler
	bt := transpiler.NewBitwiseTranspiler()

	// Analyze the example file
	result, err := bt.AnalyzeFile("examples/discord_traditional.go")
	if err != nil {
		log.Fatalf("Error analyzing file: %v", err)
	}

	// Print results
	fmt.Printf("📁 File analyzed: %s\n", result.OriginalFile)
	fmt.Printf("🎯 Optimizations found: %d\n", len(result.Optimizations))

	for i, opt := range result.Optimizations {
		fmt.Printf("  %d. %s\n", i+1, opt.Description)
		fmt.Printf("     Type: %s\n", opt.Type)
		fmt.Printf("     Bytes saved: %d\n", opt.BytesSaved)
		fmt.Printf("     Speedup: %.2fx\n", opt.SpeedupFactor)
	}

	fmt.Printf("🔒 Security features: %d\n", len(result.SecurityFeatures))
	for i, feature := range result.SecurityFeatures {
		fmt.Printf("  %d. %s (%s)\n", i+1, feature.Description, feature.Strength)
	}

	fmt.Println("✅ Analysis complete!")
}
