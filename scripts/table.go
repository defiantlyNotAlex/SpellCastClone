package main

import "fmt"

func generate_table() {
	fmt.Println("<table class=\"board\">")
	fmt.Println("    <tbody>")
	for y := range 6 {
		fmt.Println("        <tr>")
		for x := range 6 {
			fmt.Print("            <td>")
			fmt.Printf("<button id=\"cell%d%d\" onclick=\"OnClickBoard(%d, %d)\"></button>", x, y, x, y)
			fmt.Println("</td>")
		}
		fmt.Println("        </tr>")
	}
	fmt.Println("    </tbody>")
	fmt.Println("</table>")
}

func main() {
	fmt.Println("<!DOCTYPE html>")
	fmt.Println("<head lang=\"en\">")
	fmt.Println("    <title>Spell Cast Clone</title>")
	fmt.Println("</head>")
	fmt.Println("<body>")
	generate_table()
	fmt.Println("</body>")
}
