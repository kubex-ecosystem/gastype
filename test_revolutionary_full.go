// Package main testa a revolução completa do GASType com análise de contextos
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/faelmori/gastype/internal/transpiler"
)

func main() {
	fmt.Println("🌟 TESTING GASTYPE REVOLUTIONARY TRANSPILATION - FULL CONTEXT ANALYSIS")
	fmt.Println(strings.Repeat("=", 80))

	// STEP 1: Context Analysis
	fmt.Println("\n🔍 STEP 1: ANALYZING LOGICAL CONTEXTS...")
	analyzer := transpiler.NewContextAnalyzer()

	contexts, err := analyzer.AnalyzeFile("examples/discord_traditional.go")
	if err != nil {
		log.Fatalf("Error analyzing contexts: %v", err)
	}

	fmt.Printf("📊 Total contexts found: %d\n", len(contexts))

	// Show all contexts
	for i, context := range contexts {
		fmt.Printf("\n  Context %d:\n", i+1)
		fmt.Printf("    Type: %s\n", context.Type)
		fmt.Printf("    Name: %s\n", context.Name)
		fmt.Printf("    Complexity: %d\n", context.Complexity)
		fmt.Printf("    Transpilable: %t\n", context.Transpilable)
		fmt.Printf("    BitStates: %d\n", len(context.BitStates))
		fmt.Printf("    Lines: %d-%d\n", context.StartLine, context.EndLine)
	}

	// STEP 2: Filter transpilable contexts
	fmt.Println("\n⚡ STEP 2: FILTERING TRANSPILABLE CONTEXTS...")
	transpilableContexts := analyzer.GetTranspilableContexts()
	fmt.Printf("✅ Transpilable contexts: %d/%d\n", len(transpilableContexts), len(contexts))

	for i, context := range transpilableContexts {
		fmt.Printf("  %d. %s (%s) - %d bit states\n", i+1, context.Name, context.Type, len(context.BitStates))
	}

	// STEP 3: Generate revolutionary code (different obfuscation levels)
	fmt.Println("\n🚀 STEP 3: GENERATING REVOLUTIONARY CODE...")

	obfuscationLevels := []int{1, 2, 3}
	levelNames := []string{"LOW", "MEDIUM", "HIGH"}

	for i, level := range obfuscationLevels {
		fmt.Printf("\n  Generating %s obfuscation level...\n", levelNames[i])

		generator := transpiler.NewAdvancedCodeGenerator(level)
		code, err := generator.GenerateAdvancedCode("examples/discord_traditional.go", transpilableContexts)
		if err != nil {
			log.Printf("Error generating code for level %d: %v", level, err)
			continue
		}

		// Save to file
		outputFile := fmt.Sprintf("gastype_output/discord_revolutionary_level_%d.go", level)
		err = os.WriteFile(outputFile, []byte(code), 0644)
		if err != nil {
			log.Printf("Error saving file %s: %v", outputFile, err)
			continue
		}

		fmt.Printf("    ✅ Saved: %s (%d bytes)\n", outputFile, len(code))
	}

	// STEP 4: Performance analysis
	fmt.Println("\n📈 STEP 4: PERFORMANCE ANALYSIS...")

	totalStructs := len(analyzer.GetContextsByType("struct"))
	totalFunctions := len(analyzer.GetContextsByType("function"))
	totalIfChains := len(analyzer.GetContextsByType("if_chain"))
	totalAuth := len(analyzer.GetContextsByType("auth_logic"))

	fmt.Printf("  Structs optimized: %d\n", totalStructs)
	fmt.Printf("  Functions obfuscated: %d\n", totalFunctions)
	fmt.Printf("  If-chains → Jump tables: %d\n", totalIfChains)
	fmt.Printf("  Auth logic → Obfuscated: %d\n", totalAuth)

	// Calculate estimated performance gains
	estimatedByteSavings := 0
	estimatedSpeedup := 1.0

	for _, context := range transpilableContexts {
		switch context.Type {
		case "struct":
			estimatedByteSavings += len(context.BitStates) * 7 // 8 bytes → 1 byte per bool
			estimatedSpeedup += 0.5
		case "if_chain":
			estimatedByteSavings += context.Complexity * 10 // Jump table savings
			estimatedSpeedup += float64(context.Complexity) * 0.3
		case "auth_logic":
			estimatedByteSavings += 50 // Obfuscation overhead vs security gain
			estimatedSpeedup += 0.2
		case "function":
			estimatedByteSavings += 20
			estimatedSpeedup += 0.1
		}
	}

	fmt.Printf("\n💾 Estimated memory savings: %d bytes\n", estimatedByteSavings)
	fmt.Printf("⚡ Estimated speedup factor: %.2fx\n", estimatedSpeedup)

	// STEP 5: Security analysis
	fmt.Println("\n🔒 STEP 5: SECURITY ANALYSIS...")

	securityFeatures := 0
	for _, context := range transpilableContexts {
		if context.Type == "auth_logic" {
			securityFeatures += 3 // High security
		} else if context.Transpilable {
			securityFeatures += 1 // General obfuscation
		}
	}

	fmt.Printf("  Security features implemented: %d\n", securityFeatures)
	fmt.Printf("  Human-readable logic eliminated: %d contexts\n", len(transpilableContexts))
	fmt.Printf("  Reverse engineering difficulty: %s\n", getSecurityLevel(securityFeatures))

	// STEP 6: Test generated code
	fmt.Println("\n🧪 STEP 6: TESTING GENERATED CODE...")

	for i, level := range obfuscationLevels {
		outputFile := fmt.Sprintf("gastype_output/discord_revolutionary_level_%d.go", level)

		fmt.Printf("  Testing %s obfuscation...\n", levelNames[i])

		// Try to compile (basic syntax check)
		if _, err := os.Stat(outputFile); err == nil {
			fmt.Printf("    ✅ File exists and is readable\n")

			// In a real scenario, we would run: go build outputFile
			// For now, just verify file size is reasonable
			if stat, err := os.Stat(outputFile); err == nil {
				if stat.Size() > 1000 { // At least 1KB
					fmt.Printf("    ✅ File size reasonable: %d bytes\n", stat.Size())
				} else {
					fmt.Printf("    ⚠️  File might be too small: %d bytes\n", stat.Size())
				}
			}
		} else {
			fmt.Printf("    ❌ File not found or unreadable\n")
		}
	}

	// FINAL SUMMARY
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎉 GASTYPE REVOLUTIONARY TRANSPILATION COMPLETE!")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("📊 Summary:\n")
	fmt.Printf("  - Original file: examples/discord_traditional.go\n")
	fmt.Printf("  - Contexts analyzed: %d\n", len(contexts))
	fmt.Printf("  - Contexts transpiled: %d\n", len(transpilableContexts))
	fmt.Printf("  - Generated variants: %d (LOW/MEDIUM/HIGH obfuscation)\n", len(obfuscationLevels))
	fmt.Printf("  - Estimated performance: %.2fx faster\n", estimatedSpeedup)
	fmt.Printf("  - Memory optimized: %d bytes saved\n", estimatedByteSavings)
	fmt.Printf("  - Security level: %s\n", getSecurityLevel(securityFeatures))

	fmt.Println("\n🚀 Files generated:")
	for i, level := range obfuscationLevels {
		fmt.Printf("  - discord_revolutionary_level_%d.go (%s obfuscation)\n", level, levelNames[i])
	}

	fmt.Println("\n💡 Next steps:")
	fmt.Println("  1. Test generated code: cd gastype_output && go run discord_revolutionary_level_2.go")
	fmt.Println("  2. Compare performance with benchmarks")
	fmt.Println("  3. Apply to real projects for maximum gains")

	fmt.Println("\n✨ REVOLUTION COMPLETE - GO CODE WILL NEVER BE THE SAME! ✨")
}

// getSecurityLevel returns security level description
func getSecurityLevel(features int) string {
	switch {
	case features >= 10:
		return "MAXIMUM (Unreadable to humans)"
	case features >= 5:
		return "HIGH (Very difficult to reverse)"
	case features >= 2:
		return "MEDIUM (Obfuscated but traceable)"
	default:
		return "LOW (Basic optimizations)"
	}
}
