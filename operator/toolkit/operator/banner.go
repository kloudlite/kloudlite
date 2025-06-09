package operator

import (
	"fmt"
	"time"
)

var BuiltAt string

func printReadyBanner(readyIn time.Duration) {
	fmt.Printf(`
         **                                                               
       ****                                                               
      ****     *                                                          
    ****     *****       ▗▖ ▗▖▗▖    ▗▄▖ ▗▖ ▗▖▗▄▄▄ ▗▖   ▗▄▄▄▖▗▄▄▄▖▗▄▄▄▖    
  ****     *********     ▐▌▗▞▘▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌  █▐▌     █    █  ▐▌       
 ****     ************   ▐▛▚▖ ▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌  █▐▌     █    █  ▐▛▀▀▘    
  ****     *********     ▐▌ ▐▌▐▙▄▄▖▝▚▄▞▘▝▚▄▞▘▐▙▄▄▀▐▙▄▄▖▗▄█▄▖  █  ▐▙▄▄▖    
    ****     *****                                                        
      ****     *            🚀 running in %.2fs                           
       ****                                                               
         **                                                               
                                                                          
`, readyIn.Seconds())
}

func printBuildInfo() {
	if BuiltAt == "" {
		BuiltAt = time.Now().Format(time.RFC3339)
	}
	fmt.Printf(`
📦 built at %s
  `, BuiltAt)
}
