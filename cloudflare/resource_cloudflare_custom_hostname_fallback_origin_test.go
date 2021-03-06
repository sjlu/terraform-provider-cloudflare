package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccCloudflareCustomHostnameFallbackOrigin(t *testing.T) {
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_hostname_fallback_origin." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareCustomHostnameFallbackOriginDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomHostnameFallbackOrigin(zoneID, rnd, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "origin", fmt.Sprintf("fallback-origin.%s.%s", rnd, domain)),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomHostnameFallbackOrigin(zoneID, rnd, subdomain, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_custom_hostname_fallback_origin" "%[2]s" {
  zone_id = "%[1]s"
  origin = "fallback-origin.%[3]s.%[4]s"
}

resource "cloudflare_record" "%[2]s" {
  zone_id = "%[1]s"
  name    = "fallback-origin.%[2]s.%[4]s"
  value   = "example.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  depends_on = [cloudflare_custom_hostname_fallback_origin.%[2]s]
}`, zoneID, rnd, subdomain, domain)
}

func TestAccCloudflareCustomHostnameFallbackOriginUpdate(t *testing.T) {
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	rnd := generateRandomResourceName()
	rndUpdate := rnd + "-updated"
	resourceName := "cloudflare_custom_hostname_fallback_origin." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareCustomHostnameFallbackOriginDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomHostnameFallbackOrigin(zoneID, rnd, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "origin", fmt.Sprintf("fallback-origin.%s.%s", rnd, domain)),
				),
			},
			{
				Config: testAccCheckCloudflareCustomHostnameFallbackOrigin(zoneID, rnd, rndUpdate, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "origin", fmt.Sprintf("fallback-origin.%s.%s", rndUpdate, domain)),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomHostnameFallbackOriginDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_custom_hostname_fallback_origin" {
			continue
		}

		fallbackOrigin, err := client.CustomHostnameFallbackOrigin(rs.Primary.Attributes["zone_id"])

		// If the fallback origin is in the process of being deleted, that's fine to
		// say it's been deleted as the remote API will take care of it.
		if fallbackOrigin.Status != "pending_deletion" && err == nil {
			return fmt.Errorf("Fallback Origin still exists")
		}
	}

	return nil
}
