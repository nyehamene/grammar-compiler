package testutil

import "fmt"

func Grammar(rules string) []byte {
	return []byte(rules)
}

func Rule(name, body string) string {
	return name + " = " + body + ";"
}

func RuleInline(name, body string) string {
	return fmt.Sprintf("%s = %s;", name, body)
}

func WithImport(pkgName, importedPkg string) string {
	return fmt.Sprintf(`%s = @import("%s");`, pkgName, importedPkg)
}

func WithPackage(pkgName string) string {
	return fmt.Sprintf(`@package("%s");`, pkgName)
}

func RuleWithImport(name, importedPkg string) string {
	return fmt.Sprintf(`%s = @import("%s");`, name, importedPkg)
}

func Module(rules string) string {
	return rules
}

var SimpleGrammar = `
a = "hello";
b = a;
`

var GrammarWithImport = `
main = @import("lib");
lib = "hello";
`

var PackageGrammar = `
@package("mypackage");

rule_a = "hello";
`

var MultiRuleGrammar = `
rule_a = "a";
rule_b = "b";
rule_c = rule_a | rule_b;
`
