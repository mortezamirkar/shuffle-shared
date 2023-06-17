package shuffle

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"

	//"os/exec"
	mathrand "math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"

	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"

	"github.com/bradfitz/slice"
	uuid "github.com/satori/go.uuid"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/frikky/kin-openapi/openapi2"
	"github.com/frikky/kin-openapi/openapi2conv"
	"github.com/frikky/kin-openapi/openapi3"

	"github.com/google/go-github/v28/github"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine"
)

var project ShuffleStorage
var baseDockerName = "frikky/shuffle"
var SSOUrl = ""

func GetUsecaseData() string {
	return (`[
    {
        "name": "1. Collect",
        "color": "#c51152",
        "list": [
            {
                "name": "Email management",
								"priority": 100,
								"type": "communication",
								"last": "cases", 
                "items": {
                    "name": "Release a quarantined message",
                    "items": {}
                }
            },
            {
                "name": "EDR to ticket",
								"priority": 100,
								"type": "edr",
								"last": "cases",
                "items": {
                    "name": "Get host information",
                    "items": {}
                }
            },
            {
                "name": "SIEM to ticket",
								"priority": 100,
								"type": "siem",
								"last": "cases",
								"description": "Ensure tickets are forwarded to the correct destination. Alternatively add enrichment on it's way there.",
								"video": "https://www.youtube.com/watch?v=FBISHA7V15c&t=197s&ab_channel=OpenSecure",
								"blogpost": "https://medium.com/shuffle-automation/introducing-shuffle-an-open-source-soar-platform-part-1-58a529de7d12",
								"reference_image": "/images/detectionframework.png",
                "items": {}
            },
            {
                "name": "2-way Ticket synchronization",
								"priority": 60,
                "items": {}
            },
            {
                "name": "ChatOps",
								"priority": 70,
                "items": {}
            },
            {
                "name": "Threat Intel received",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Assign tickets",
								"priority": 30,
                "items": {}
            },
            {
                "name": "Firewall alerts",
								"priority": 90,
								"type": "network",
								"last": "cases",
                "items": {
                    "name": "URL filtering",
                    "items": {}
                }
            },
            {
                "name": "IDS/IPS alerts",
								"type": "network",
								"last": "cases",
								"priority": 90,
                "items": {
                    "name": "Manage policies",
                    "items": {}
                }
            },
            {
                "name": "Deduplicate information",
								"priority": 70,
                "items": {}
            }
        ]
    },
    {
        "name": "2. Enrich",
        "color": "#f4c20d",
        "list": [
            {
                "name": "Internal Enrichment",
								"priority": 100,
								"type": "intel",
                "items": {
                    "name": "...",
                    "items": {}
                }
            },
            {
                "name": "External historical Enrichment",
								"priority": 90,
								"type": "intel",
                "items": {
                    "name": "...",
                    "items": {}
                }
            },
            {
                "name": "Sandbox",
								"priority": 60,
								"type": "intel",
                "items": {
                    "name": "Use a sandbox to analyze",
                    "items": {}
                }
            },
            {
                "name": "Realtime",
								"priority": 50,
								"type": "intel",
                "items": {
                    "name": "Analyze screenshots, websites etc. in realtime",
                    "items": {}
                }
						}
        ]
    },
    {
        "name": "3. Detect",
        "color": "#3cba54",
        "list": [
            {
                "name": "Search SIEM (Sigma)",
								"priority": 90,
								"type": "siem",
                "items": {
                    "name": "Endpoint",
                    "items": {}
                }
            },
            {
                "name": "Search EDR (OSQuery)",
								"type": "edr",
								"priority": 90,
                "items": {}
            },
            {
                "name": "Search emails (Sublime)",
								"priority": 90,
								"type": "communication",
                "items": {
                    "name": "Check headers and IOCs",
                    "items": {}
                }
            },
            {
                "name": "Search IOCs (ioc-finder)",
								"priority": 50,
								"type": "intel",
                "items": {}
            },
            {
                "name": "Search files (Yara)",
								"priority": 50,
								"type": "intel",
                "items": {}
            },
            {
                "name": "Memory Analysis (Volatility)",
								"priority": 50,
								"type": "intel",
                "items": {}
            },
            {
                "name": "IDS & IPS (Snort/Surricata)",
								"priority": 50,
								"type": "network",
                "items": {}
            },
            {
                "name": "Honeypot access",
								"priority": 50,
								"type": "network",
                "items": {
                    "name": "...",
                    "items": {}
                }
            }
        ]
    },
    {
        "name": "4. Respond",
        "color": "#4885ed",
        "list": [
            {
                "name": "Eradicate malware",
								"priority": 90,
								"type": "intel",
                "items": {}
            },
            {
                "name": "Quarantine host(s)",
								"priority": 90,
								"type": "edr",
                "items": {}
            },
            {
                "name": "Block IPs, URLs, Domains and Hashes",
								"priority": 90,
								"type": "network",
                "items": {}
            },
            {
                "name": "Trigger scans",
								"priority": 50,
								"type": "assets",
                "items": {}
            },
            {
                "name": "Update indicators (FW, EDR, SIEM...)",
								"priority": 50,
								"type": "intel",
                "items": {}
            },
            {
                "name": "Autoblock activity when threat intel is received",
								"priority": 50,
								"type": "intel",
								"last": "iam",
                "items": {}
            },
            {
                "name": "Lock/Delete/Reset account",
								"priority": 50,
								"type": "iam",
                "items": {}
            },
            {
                "name": "Lock vault",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Increase authentication",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Get policies from assets",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Run ansible scripts",
								"priority": 50,
                "items": {}
            }
        ]
    },
    {
        "name": "5. Verify",
        "color": "#7f00ff",
        "list": [
            {
                "name": "Discover vulnerabilities",
								"priority": 80,
								"type": "assets",
                "items": {}
            },
            {
                "name": "Discover assets",
								"priority": 80,
								"type": "assets",
                "items": {}
            },
            {
                "name": "Ensure policies are followed",
								"priority": 80,
                "items": {}
            },
            {
                "name": "Find Inactive users",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Botnet tracker",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Ensure access rights match HR systems",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Ensure onboarding is followed",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Third party apps in SaaS",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Devices used for your cloud account",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Too much access in GCP/Azure/AWS/ other clouds",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Certificate validation",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Monitor domain creation and expiration",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Monitor new DNS entries for domain with passive DNS",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Monitor and track password dumps",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Monitor for mentions of domain on darknet sites",
								"priority": 50,
                "items": {}
            },
            {
                "name": "Reporting",
								"priority": 50,
								"keywords": ["report", "reporting", "sheets", "excel"],
								"keyword_matches": 1,
                "items": {
                    "name": "Monthly reports",
                    "items": {
                        "name": "...",
                        "items": {}
                    }
                }
            }
        ]
    }
]`)
}

var sandboxProject = "shuffle-sandbox-337810"

func GetContext(request *http.Request) context.Context {
	if project.Environment == "cloud" && len(memcached) == 0 {
		return appengine.NewContext(request)
	}

	return context.Background()
}

func HandleCors(resp http.ResponseWriter, request *http.Request) bool {

	// FIXME - this is to handle multiple frontends in test rofl
	origin := request.Header["Origin"]
	resp.Header().Set("Vary", "Origin")
	if len(origin) > 0 {
		resp.Header().Set("Access-Control-Allow-Origin", origin[0])
		//resp.Header().Set("Access-Control-Allow-Origin", "https://ca.shuffler.io")
		//resp.Header().Set("Access-Control-Allow-Origin", "http://localhost:3002")
	} else {
		resp.Header().Set("Access-Control-Allow-Origin", "http://localhost:4201")
	}
	//resp.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, X-Requested-With, remember-me")
	resp.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, POST, PATCH")
	resp.Header().Set("Access-Control-Allow-Credentials", "true")

	if request.Method == "OPTIONS" {
		resp.WriteHeader(200)
		resp.Write([]byte("OK"))
		return true
	}

	return false
}

func Md5sum(data []byte) string {
	hasher := md5.New()
	hasher.Write(data)
	newmd5 := hex.EncodeToString(hasher.Sum(nil))

	return newmd5
}

func Md5sumfile(filepath string) string {
	dat, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Printf("Error in dat: %s", err)
	}

	hasher := md5.New()
	hasher.Write(dat)
	newmd5 := hex.EncodeToString(hasher.Sum(nil))

	log.Printf("%s: %s", filepath, newmd5)
	return newmd5
}

func randStr(strSize int, randType string) string {

	var dictionary string

	if randType == "alphanum" {
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "alpha" {
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "number" {
		dictionary = "0123456789"
	}

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(bytes)
}

func HandleSet2fa(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[AUDIT] Api authentication failed in get 2fa: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("[ERROR] Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error with body read: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	type parsedValue struct {
		Code   string `json:"code"`
		UserId string `json:"user_id"`
	}

	var tmpBody parsedValue
	err = json.Unmarshal(body, &tmpBody)
	if err != nil {
		log.Printf("[WARNING] Error with unmarshal tmpBody in verify 2fa: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if len(tmpBody.Code) != 6 {
		log.Printf("[WARNING] Length of code isn't 6: %s", tmpBody.Code)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Length of code must be 6"}`)))
		return
	}

	// FIXME: Everything should match?
	// || user.Id != tmpBody.UserId
	if user.Id != fileId {
		log.Printf("[WARNING] Bad ID: %s vs %s", user.Id, fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can only set 2fa for your own user. Pass field user_id in JSON."}`)))
		return
	}

	ctx := GetContext(request)
	foundUser, err := GetUser(ctx, user.Id)
	if err != nil {
		log.Printf("[ERROR] Can't find user %s (set 2fa): %s", user.Id, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed getting your user."}`)))
		return
	}

	//https://www.gojek.io/blog/a-diy-two-factor-authenticator-in-golang
	interval := time.Now().Unix() / 30
	HOTP, err := getHOTPToken(foundUser.MFA.PreviousCode, interval)
	if err != nil {
		log.Printf("[ERROR] Failed generating a HOTP token: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if HOTP != tmpBody.Code {
		log.Printf("[DEBUG] Bad code sent for user %s (%s). Sent: %s, Want: %s", user.Username, user.Id, tmpBody.Code, HOTP)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Wrong code. Try again"}`))
		return
	}

	foundUser.MFA.Active = true
	foundUser.MFA.ActiveCode = foundUser.MFA.PreviousCode
	foundUser.MFA.PreviousCode = ""
	err = SetUser(ctx, foundUser, true)
	if err != nil {
		log.Printf("[WARNING] Failed SETTING MFA for user %s (%s): %s", foundUser.Username, foundUser.Id, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed updating your user. Please try again."}`))
		return
	}

	log.Printf("[DEBUG] Successfully enabled 2FA for user %s (%s)", foundUser.Username, foundUser.Id)

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true, "reason": "Correct code. MFA is now required for this user."}`))
}

func getHOTPToken(secret string, interval int64) (string, error) {

	//Converts secret to base32 Encoding. Base32 encoding desires a 32-character
	//subset of the twenty-six letters A–Z and ten digits 0–9
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", err
	}

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(interval))

	//Signing the value using HMAC-SHA1 Algorithm
	hash := hmac.New(sha1.New, key)
	hash.Write(bs)
	h := hash.Sum(nil)

	// We're going to use a subset of the generated hash.
	// Using the last nibble (half-byte) to choose the index to start from.
	// This number is always appropriate as it's maximum decimal 15, the hash will
	// have the maximum index 19 (20 bytes of SHA1) and we need 4 bytes.
	o := (h[19] & 15)

	var header uint32
	//Get 32 bit chunk from hash starting at the o
	r := bytes.NewReader(h[o : o+4])
	err = binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return "", err
	}

	//Ignore most significant bits as per RFC 4226.
	//Takes division from one million to generate a remainder less than < 7 digits
	h12 := (int(header) & 0x7fffffff) % 1000000

	//Converts number as a string
	otp := strconv.Itoa(int(h12))

	// Dumb solutions <3
	// This works well, as the numbers are small ^_^
	if len(otp) == 0 {
		otp = "000000"
	} else if len(otp) == 1 {
		otp = fmt.Sprintf("00000%s", otp)
	} else if len(otp) == 2 {
		otp = fmt.Sprintf("0000%s", otp)
	} else if len(otp) == 3 {
		otp = fmt.Sprintf("000%s", otp)
	} else if len(otp) == 4 {
		otp = fmt.Sprintf("00%s", otp)
	} else if len(otp) == 5 {
		otp = fmt.Sprintf("0%s", otp)
	}

	return otp, nil
}

func HandleGet2fa(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[ERROR] Api authentication failed in get 2fa: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("[ERROR] Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if user.Id != fileId {
		log.Printf("[WARNING] Bad ID: %s vs %s", user.Id, fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can only set 2fa for your own user"}`)))
		return
	}

	// https://socketloop.com/tutorials/golang-generate-qr-codes-for-google-authenticator-app-and-fix-cannot-interpret-qr-code-error

	// generate a random string - preferably 6 or 8 characters
	randomStr := randStr(8, "alphanum")

	// For Google Authenticator purpose
	// for more details see
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	secret := base32.StdEncoding.EncodeToString([]byte(randomStr))

	// authentication link. Remember to replace SocketLoop with yours.
	// for more details see
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	authLink := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=Shuffle", user.Username, secret)
	png, err := qrcode.Encode(authLink, qrcode.Medium, 256)
	if err != nil {
		log.Printf("[ERROR] Failed PNG encoding: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed image encoding"}`)))
		return
	}

	dataURI := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString([]byte(png)))
	newres := ResultChecker{
		Success: true,
		Reason:  dataURI,
		Extra:   strings.ReplaceAll(secret, "=", "A"),
	}

	newjson, err := json.Marshal(newres)
	if err != nil {
		log.Printf("[ERROR] Failed marshal in get OTP: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking data"}`)))
		return
	}

	ctx := GetContext(request)
	//user.MFA.PreviousCode = authLink
	user.MFA.PreviousCode = secret
	err = SetUser(ctx, &user, true)
	if err != nil {
		log.Printf("[WARNING] Failed updating MFA for user: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed updating your user"}`))
		return
	}

	log.Printf("[DEBUG] Sent new MFA update for user %s (%s)", user.Username, user.Id)
	//log.Printf("%s", newjson)

	resp.WriteHeader(200)
	resp.Write([]byte(newjson))
}

func HandleGetOrgs(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting GET ORGS request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get orgs: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)
	if user.Role != "global_admin" {
		orgs := []OrgMini{}
		for _, item := range user.Orgs {
			// FIXM: Should return normal orgs, but hidden if the user isn't admin
			org, err := GetOrg(ctx, item)
			if err == nil {
				orgs = append(orgs, OrgMini{
					Id:         org.Id,
					Name:       org.Name,
					CreatorOrg: org.CreatorOrg,
					Image:      org.Image,
				})
				// Role:       "admin",
			}
		}

		newjson, err := json.Marshal(orgs)
		if err != nil {
			log.Printf("[WARNING] Failed marshal in get orgs: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking"}`)))
			return
		}

		//log.Printf("[AUDIT] User %s (%s) isn't global admin and can't list orgs. Returning list of local orgs.", user.Username, user.Id)
		resp.WriteHeader(200)
		resp.Write([]byte(newjson))
		return
	}

	orgs, err := GetAllOrgs(ctx)
	if err != nil || len(orgs) == 0 {
		log.Printf("[WARNING] Failed getting orgs: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Can't get orgs"}`))
		return
	}

	newjson, err := json.Marshal(orgs)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshal in get orgs: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking"}`)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func HandleGetOrg(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Checking if it's a special region. All user-specific requests should
	// go through shuffler.io and not subdomains
	if project.Environment == "cloud" {
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting GET ORG request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if strings.Contains(fileId, "?") {
		fileId = strings.Split(fileId, "?")[0]
	}

	ctx := GetContext(request)
	sanitizeOrg := false
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {

		// This is specifically for public workflows
		referenceId, referenceok := request.URL.Query()["reference_execution"]
		authorization, authorizationok := request.URL.Query()["authorization"]
		if referenceok && authorizationok {
			workflowExecution, err := GetWorkflowExecution(ctx, referenceId[0])
			if err == nil && authorization[0] == workflowExecution.Authorization {
				sanitizeOrg = true
			}
		}

		if sanitizeOrg != true {
			log.Printf("[WARNING] Api authentication failed in get org: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	org, err := GetOrg(ctx, fileId)
	if err != nil {
		log.Printf("[WARNING] Failed getting org '%s': %s", fileId, err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Failed getting org details"}`))
		return
	}

	admin := false
	if user.SupportAccess == true {
		admin = true
		sanitizeOrg = false

		// Update active org for user to this one?
		// This makes it possible to walk around in the UI for the org
		if user.ActiveOrg.Id != org.Id {
			log.Printf("[AUDIT] User %s (%s) is admin and has access to org %s. Updating active org to this one.", user.Username, user.Id, org.Id)
			user.ActiveOrg.Id = org.Id
			user.ActiveOrg.Name = org.Name
			user.Role = "admin"

			SetUser(ctx, &user, false)

			DeleteCache(ctx, fmt.Sprintf("%s_workflows", user.Id))
			DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
			DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
			DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))
		}

	} else {
		userFound := false
		for _, inneruser := range org.Users {
			if inneruser.Id == user.Id {
				userFound = true

				if inneruser.Role == "admin" {
					admin = true
				}

				break
			}
		}

		if !userFound && !sanitizeOrg {
			log.Printf("[ERROR] User '%s' (%s) isn't a part of org %s (get)", user.Username, user.Id, org.Id)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "User doesn't have access to org"}`))
			return

		}
	}

	if !admin {
		org.Defaults = Defaults{}
		org.SSOConfig = SSOConfig{}
		org.Subscriptions = []PaymentSubscription{}
		org.ManagerOrgs = []OrgMini{}
		org.ChildOrgs = []OrgMini{}
		org.Invites = []string{}
	} else {
		org.SyncFeatures.AppExecutions.Description = "The amount of Apps within Workflows you can run per month. This limit can be exceeded when running workflows without a trigger (manual execution)."
		org.SyncFeatures.WorkflowExecutions.Description = "N/A. See App Executions"
		org.SyncFeatures.Webhook.Description = "Webhooks are Triggers that take an HTTP input to start a workflow. Read docs for more."
		org.SyncFeatures.Schedules.Description = "Schedules are Triggers that run on an interval defined by you. Read docs for more."
		org.SyncFeatures.MultiEnv.Description = "Multiple Environments are used to run automation in different physical locations. Change from /admin?tab=environments"
		org.SyncFeatures.MultiTenant.Description = "Multiple Tenants can be used to segregate information for each MSSP Customer. Change from /admin?tab=suborgs"
		//org.SyncFeatures.MultiTenant.Description = "Multiple Tenants can be used to segregate information for each MSSP Customer. Change from /admin?tab=suborgs"

		//log.Printf("LIMIT: %s", org.SyncFeatures.AppExecutions.Limit)
		orgChanged := false
		if org.SyncFeatures.AppExecutions.Limit == 0 || org.SyncFeatures.AppExecutions.Limit == 1500 {
			org.SyncFeatures.AppExecutions.Limit = 10000
			orgChanged = true
		}

		if org.SyncFeatures.SendMail.Limit == 0 {
			org.SyncFeatures.SendMail.Limit = 100
			orgChanged = true
		}

		if org.SyncFeatures.SendSms.Limit == 0 {
			org.SyncFeatures.SendSms.Limit = 30
			orgChanged = true
		}

		org.SyncFeatures.EmailTrigger.Limit = 0
		if org.SyncFeatures.MultiEnv.Limit == 0 {
			org.SyncFeatures.MultiEnv.Limit = 1
			orgChanged = true
		}

		org.SyncFeatures.EmailTrigger.Limit = 0

		if orgChanged {
			log.Printf("[DEBUG] Org features for %s (%s) changed. Updating.", org.Name, org.Id)
			err = SetOrg(ctx, *org, org.Id)
			if err != nil {
				log.Printf("[WARNING] Failed updating org during org loading")
			}
		}

		info, err := GetOrgStatistics(ctx, fileId)
		if err == nil {
			org.SyncFeatures.AppExecutions.Usage = info.MonthlyAppExecutions
		}

		org.SyncFeatures.MultiTenant.Usage = int64(len(org.ChildOrgs) + 1)
		envs, err := GetEnvironments(ctx, fileId)
		if err == nil {
			//log.Printf("Envs: %s", len(envs))
			org.SyncFeatures.MultiEnv.Usage = int64(len(envs))
		}
	}

	// Make sure to add all orgs that are childs if you have access
	org.ChildOrgs = []OrgMini{}
	for _, orgloop := range user.Orgs {
		suborg, err := GetOrg(ctx, orgloop)
		if err != nil {
			continue
		}

		// Check if current user is in that org
		found := false
		for _, userloop := range suborg.Users {
			if userloop.Id == user.Id {
				found = true
			}
		}

		if !found {
			continue
		}

		if suborg.CreatorOrg == org.Id {
			org.ChildOrgs = append(org.ChildOrgs, OrgMini{
				Id:         suborg.Id,
				Name:       suborg.Name,
				CreatorOrg: suborg.CreatorOrg,
				Image:      suborg.Image,
				RegionUrl:  suborg.RegionUrl,
			})
		}
	}

	org.Users = []User{}
	org.SyncConfig.Apikey = ""
	org.SyncConfig.Source = ""

	// This is for sending branding information
	// to those who need it
	if sanitizeOrg {
		newOrg := org
		org = &Org{}
		org.Name = newOrg.Name
		org.Id = newOrg.Id
		org.Image = newOrg.Image
		org.RegionUrl = newOrg.RegionUrl
		org.Org = newOrg.Org

	}

	if !user.SupportAccess {
		org.LeadInfo = LeadInfo{}
	}

	newjson, err := json.Marshal(org)
	if err != nil {
		log.Printf("[ERROR] Failed unmarshal of org %s (%s): %s", org.Name, org.Id, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking"}`)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func HandleLogout(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	ctx := GetContext(request)

	runReturn := false
	userInfo, usererr := HandleApiAuthentication(resp, request)
	log.Printf("[AUDIT] Logging out user %s (%s)", userInfo.Username, userInfo.Id)
	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting LOGOUT request to main site handler (shuffler.io)")
			DeleteCache(ctx, fmt.Sprintf("%s_workflows", userInfo.Id))
			DeleteCache(ctx, fmt.Sprintf("apps_%s", userInfo.Id))
			DeleteCache(ctx, fmt.Sprintf("user_%s", strings.ToLower(userInfo.Username)))
			DeleteCache(ctx, fmt.Sprintf("user_%s", userInfo.Id))
			DeleteCache(ctx, fmt.Sprintf("session_%s", userInfo.Session))

			RedirectUserRequest(resp, request)

			DeleteCache(ctx, fmt.Sprintf("%s_workflows", userInfo.Id))
			DeleteCache(ctx, fmt.Sprintf("apps_%s", userInfo.Id))
			DeleteCache(ctx, fmt.Sprintf("user_%s", strings.ToLower(userInfo.Username)))
			DeleteCache(ctx, fmt.Sprintf("user_%s", userInfo.Id))
			DeleteCache(ctx, fmt.Sprintf("session_%s", userInfo.Session))

			// FIXME: Allow superfluous cleanups?
			// Point is: should it continue running the logout to
			// ensure cookies are cleared?
			// Keeping it for now to ensure cleanup.
			return
		}
	}

	c, err := request.Cookie("session_token")
	if err != nil {
		c, err = request.Cookie("__session")
	}

	if err == nil {
		newCookie := &http.Cookie{
			Name:    "session_token",
			Value:   c.Value,
			Expires: time.Now().Add(-100 * time.Hour),
			MaxAge:  -1,
		}
		if project.Environment == "cloud" {
			newCookie.Domain = ".shuffler.io"
			newCookie.Secure = true
			newCookie.HttpOnly = true
		}

		http.SetCookie(resp, newCookie)

		newCookie.Name = "__session"
		http.SetCookie(resp, newCookie)

	} else {
		newCookie := &http.Cookie{
			Name:    "session_token",
			Value:   "",
			Expires: time.Now().Add(-100 * time.Hour),
			MaxAge:  -1,
		}

		if project.Environment == "cloud" {
			newCookie.Domain = ".shuffler.io"
			newCookie.Secure = true
			newCookie.HttpOnly = true
		}

		http.SetCookie(resp, newCookie)

		newCookie.Name = "__session"
		http.SetCookie(resp, newCookie)
	}

	DeleteCache(ctx, fmt.Sprintf("%s_workflows", userInfo.Id))
	DeleteCache(ctx, fmt.Sprintf("apps_%s", userInfo.Id))
	if runReturn == true {
		DeleteCache(ctx, fmt.Sprintf("user_%s", strings.ToLower(userInfo.Username)))
		DeleteCache(ctx, fmt.Sprintf("session_%s", userInfo.Session))
		DeleteCache(ctx, userInfo.Session)

		log.Printf("[INFO] Returning from logout request after cache cleanup")

		return
	}

	if usererr != nil {
		log.Printf("[WARNING] Api authentication failed in handleLogout: %s", err)
		resp.WriteHeader(200)
		resp.Write([]byte(`{"success": true, "reason": "Not logged in"}`))
		return
	}

	DeleteCache(ctx, fmt.Sprintf("user_%s", strings.ToLower(userInfo.Username)))
	DeleteCache(ctx, fmt.Sprintf("session_%s", userInfo.Session))
	DeleteCache(ctx, userInfo.Session)

	userInfo.Session = ""
	err = SetUser(ctx, &userInfo, true)
	if err != nil {
		log.Printf("Failed updating user: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed updating apikey"}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": false, "reason": "Successfully logged out"}`))
}

// A search for apps based on name and such
// This was before Algolia
func GetSpecificApps(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Just need to be logged in
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in set new app: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Used for searching
	returnData := fmt.Sprintf(`{"success": true, "reason": []}`)
	resp.WriteHeader(200)
	resp.Write([]byte(returnData))
	return

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error with body read: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	type tmpStruct struct {
		Search string `json:"search"`
	}

	var tmpBody tmpStruct
	err = json.Unmarshal(body, &tmpBody)
	if err != nil {
		log.Printf("[WARNING] Error with unmarshal tmpBody specific apps: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// FIXME - continue the search here with github repos etc.
	// Caching might be smart :D
	ctx := GetContext(request)
	workflowapps, err := GetPrioritizedApps(ctx, user)
	if err != nil {
		log.Printf("[WARNING] Error: Failed getting workflowapps: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	returnValues := []WorkflowApp{}
	search := strings.ToLower(tmpBody.Search)
	for _, app := range workflowapps {
		if !app.Activated && app.Generated {
			// This might be heavy with A LOT
			// Not too worried with todays tech tbh..
			appName := strings.ToLower(app.Name)
			appDesc := strings.ToLower(app.Description)
			if strings.Contains(appName, search) || strings.Contains(appDesc, search) {
				//log.Printf("Name: %s, Generated: %s, Activated: %s", app.Name, strconv.FormatBool(app.Generated), strconv.FormatBool(app.Activated))
				returnValues = append(returnValues, app)
			}
		}
	}

	newbody, err := json.Marshal(returnValues)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking workflow executions"}`)))
		return
	}

	returnData = fmt.Sprintf(`{"success": true, "reason": %s}`, string(newbody))
	resp.WriteHeader(200)
	resp.Write([]byte(returnData))
}

func GetAppAuthentication(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, userErr := HandleApiAuthentication(resp, request)
	if userErr != nil {
		log.Printf("[AUDIT] Api authentication failed in get app auth: %s", userErr)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)
	allAuths, err := GetAllWorkflowAppAuth(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get all app auth: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if len(allAuths) == 0 {
		resp.WriteHeader(200)
		resp.Write([]byte(`{"success": true, "data": []}`))
		return
	}

	// Cleanup for frontend
	newAuth := []AppAuthenticationStorage{}
	for _, auth := range allAuths {
		newAuthField := auth
		for index, _ := range auth.Fields {
			newAuthField.Fields[index].Value = "Secret. Replaced during app execution!"
		}

		newAuth = append(newAuth, newAuthField)
	}

	type returnStruct struct {
		Success bool                       `json:"success"`
		Data    []AppAuthenticationStorage `json:"data"`
	}

	allAuth := returnStruct{
		Success: true,
		Data:    allAuths,
	}

	newbody, err := json.Marshal(allAuth)
	if err != nil {
		log.Printf("Failed unmarshalling all app auths: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking workflow app auth"}`)))
		return
	}

	//data := fmt.Sprintf(`{"success": true, "data": %s}`, string(newbody))

	resp.WriteHeader(200)
	resp.Write([]byte(newbody))
}

func AddAppAuthentication(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, userErr := HandleApiAuthentication(resp, request)
	if userErr != nil {
		log.Printf("[WARNING] Api authentication failed in add app auth: %s", userErr)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to set new workflowapp: %s (%s)", user.Username, user.Id)
		resp.WriteHeader(403)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	log.Printf("[AUDIT] Setting new authentication for user %s (%s)", user.Username, user.Id)

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Error with body read in new app auth: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var appAuth AppAuthenticationStorage
	err = json.Unmarshal(body, &appAuth)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshaling (appauth): %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)
	if len(appAuth.Id) == 0 {
		appAuth.Id = uuid.NewV4().String()
	} else {
		auth, err := GetWorkflowAppAuthDatastore(ctx, appAuth.Id)
		if err == nil {
			// OrgId         string                `json:"org_id" datastore:"org_id"`
			if auth.OrgId != user.ActiveOrg.Id {
				log.Printf("[WARNING] User isn't a part of the right org during auth edit")
				resp.WriteHeader(409)
				resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": ":("}`)))
				return
			}

			if user.Role != "admin" {
				log.Printf("[AUDIT] User isn't admin during auth edit")
				resp.WriteHeader(409)
				resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": ":("}`)))
				return
			}

			if !auth.Active {
				log.Printf("[WARNING] Auth isn't active for edit")
				resp.WriteHeader(409)
				resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can't update an inactive auth"}`)))
				return
			}

			if auth.App.Name != appAuth.App.Name {
				log.Printf("[WARNING] User tried to modify auth")
				resp.WriteHeader(409)
				resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Bad app configuration: need to specify correct name"}`)))
				return
			}

			// Setting this to ensure that any new config is encrypted anew
			auth.Encrypted = false
		}
	}

	if len(appAuth.Label) == 0 {
		resp.WriteHeader(409)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Label can't be empty"}`)))
		return
	}

	// Super basic check
	if len(appAuth.App.ID) != 36 && len(appAuth.App.ID) != 32 {
		log.Printf("[WARNING] Bad ID for app: %s", appAuth.App.ID)
		resp.WriteHeader(409)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "App has to be defined"}`)))
		return
	}

	app, err := GetApp(ctx, appAuth.App.ID, user, false)
	if err != nil {
		log.Printf("[DEBUG] Failed finding app %s (%s) while setting auth. Finding it by looping apps.", appAuth.App.Name, appAuth.App.ID)
		workflowapps, err := GetPrioritizedApps(ctx, user)
		if err != nil {
			resp.WriteHeader(409)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
			return
		}

		foundIndex := -1
		for i, workflowapp := range workflowapps {
			if workflowapp.Name == appAuth.App.Name {
				foundIndex = i
				break
			}
		}

		if foundIndex >= 0 {
			log.Printf("[INFO] Found app %s (%s) by looping auth with %d parameters", workflowapps[foundIndex].Name, workflowapps[foundIndex].ID, len(workflowapps[foundIndex].Authentication.Parameters))
			app = &workflowapps[foundIndex]
			//appAuth.App.Name, appAuth.App.ID, len(appAuth.Fields)))
		} else {
			log.Printf("[ERROR] Failed finding app %s which has auth after looping", appAuth.App.ID)
			resp.WriteHeader(409)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed finding app %s (%s)"}`, appAuth.App.Name, appAuth.App.ID)))
			return
		}
	} else {
		org, err := GetOrg(ctx, user.ActiveOrg.Id)
		if err != nil {
			log.Printf("[WARNING] Failed getting org %s during app auth: %s", user.ActiveOrg.Id, err)
		} else {
			if !ArrayContains(org.ActiveApps, app.ID) {
				org.ActiveApps = append(org.ActiveApps, app.ID)
				err = SetOrg(ctx, *org, org.Id)
				if err != nil {
					log.Printf("[WARNING] Failed setting app %s for org %s during appauth", org.Id)
				} else {
					DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
					DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-100"))
					DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-500"))
					DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-1000"))
					DeleteCache(ctx, "all_apps")
					DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
					DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))
				}
			} else {
				log.Printf("[INFO] Org %s already has app %s active.", user.ActiveOrg.Id, app.ID)
			}
		}
	}

	//log.Printf("[INFO] TYPE: %s", appAuth.Type)
	if appAuth.Type == "oauth2" {
		log.Printf("[DEBUG] OAUTH2 for workflow %s. User: %s (%s)", appAuth.ReferenceWorkflow, user.Username, user.Id)

		if len(appAuth.ReferenceWorkflow) > 0 {
			workflow, err := GetWorkflow(ctx, appAuth.ReferenceWorkflow)
			if err != nil {
				log.Printf("[WARNING] WorkflowId %s doesn't exist (set oauth2)", appAuth.ReferenceWorkflow)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false}`))
				return
			}

			if user.Id != workflow.Owner || len(user.Id) == 0 {
				if workflow.OrgId == user.ActiveOrg.Id && user.Role == "admin" {
					log.Printf("[AUDIT] User %s is accessing workflow '%s' as admin (set oauth2)", user.Username, workflow.ID)
				} else if workflow.Public {
					log.Printf("[AUDIT] Letting user %s access workflow %s FOR AUTH because it's public", user.Username, workflow.ID)
				} else {
					log.Printf("[AUDIT] Wrong user (%s) for workflow %s (set oauth2)", user.Username, workflow.ID)
					resp.WriteHeader(401)
					resp.Write([]byte(`{"success": false}`))
					return
				}
			}

			// Finding count in same workflow & setting large image if missing
			count := 0
			for _, action := range workflow.Actions {
				if action.AppName == appAuth.App.Name {
					count += 1

					if len(appAuth.App.LargeImage) == 0 && len(action.LargeImage) > 0 {
						appAuth.App.LargeImage = action.LargeImage
					}

				}
			}

			appAuth.NodeCount = int64(count)
			appAuth.WorkflowCount = 1
		}

		_, err = RunOauth2Request(ctx, user, appAuth, false)
		if err != nil {
			parsederror := strings.Replace(fmt.Sprintf("%s", err), "\"", "\\\"", -1)
			log.Printf("[WARNING] Failed oauth2 request (3): %s", err)

			if strings.Contains(fmt.Sprintf("%s", err), "not consented") {
				log.Printf("Return the user to the URL with admin consent")
			}

			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed authorization: %s"}`, parsederror)))
			return
		}

		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf(`{"success": true, "reason": "Successfully set up authentication", "id": "%s"}`, appAuth.Id)))
		return
	}

	// Check if the items are correct
	for _, field := range appAuth.Fields {
		found := false
		for _, param := range app.Authentication.Parameters {
			//log.Printf("Fields: %s - %s", field, param.Name)
			if field.Key == param.Name {
				found = true
			}
		}

		if !found {
			log.Printf("[WARNING] Failed finding field %s in appauth fields for %s", field.Key, appAuth.App.Name)
			resp.WriteHeader(409)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "All auth fields required"}`)))
			return
		}
	}

	// FIXME: encryption
	//for _, param := range appAuth.Fields {
	//}

	//appAuth.LargeImage = ""
	appAuth.OrgId = user.ActiveOrg.Id
	appAuth.Defined = true
	err = SetWorkflowAppAuthDatastore(ctx, appAuth, appAuth.Id)
	if err != nil {
		log.Printf("[WARNING] Failed setting up app auth %s: %s", appAuth.Id, err)
		resp.WriteHeader(409)

		resultData := ResultChecker{
			Success: false,
			Reason:  fmt.Sprintf("%s", err),
		}

		newjson, err := json.Marshal(resultData)
		if err != nil {
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
		} else {
			resp.Write(newjson)
		}

		return
	}

	log.Printf("[INFO] Set new workflow auth for %s (%s) with ID %s", app.Name, app.ID, appAuth.Id)
	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true, "id": "%s"}`, appAuth.Id)))
}

func DeleteAppAuthentication(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, userErr := HandleApiAuthentication(resp, request)
	if userErr != nil {
		log.Printf("[WARNING] Api authentication failed in delete app auth: %s", userErr)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role != "admin" {
		log.Printf("[WARNING] Need to be admin to delete appauth")
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 5 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[5]
	}

	ctx := GetContext(request)
	nameKey := "workflowappauth"
	auth, err := GetWorkflowAppAuthDatastore(ctx, fileId)
	if err != nil {
		// Deleting cache here, as it seems to be a constant issue
		cacheKey := fmt.Sprintf("%s_%s", nameKey, user.ActiveOrg.Id)
		DeleteCache(ctx, cacheKey)

		log.Printf("[WARNING] Authget error (DELETE): %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": ":("}`))
		return
	}

	if auth.OrgId != user.ActiveOrg.Id {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "User can't edit this org"}`))
		return
	}

	// FIXME: Set affected workflows to have errors
	// 1. Get the auth
	// 2. Loop the workflows (.Usage) and set them to have errors
	// 3. Loop the nodes in workflows and do the same
	err = DeleteKey(ctx, nameKey, fileId)
	if err != nil {
		log.Printf("Failed deleting workflowapp")
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed deleting workflow app"}`)))
		return
	}

	cacheKey := fmt.Sprintf("%s_%s", nameKey, user.ActiveOrg.Id)
	DeleteCache(ctx, cacheKey)
	cacheKey = fmt.Sprintf("%s_%s", nameKey, fileId)
	DeleteCache(ctx, cacheKey)

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

func HandleSetEnvironments(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Only admin can change environments, but if there are no users, anyone can make (first)
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Can't handle set env auth"}`))
		return
	}

	if user.Role != "admin" {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Can't set environment without being admin"}`))
		return
	}

	ctx := GetContext(request)
	environments, err := GetEnvironments(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Failed getting environments: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Can't get environments when setting"}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Failed reading environment body: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed to read data"}`)))
		return
	}

	var newEnvironments []Environment
	err = json.Unmarshal(body, &newEnvironments)
	if err != nil {
		log.Printf("[ERROR] Failed unmarshaling: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed to unmarshal data"}`)))
		return
	}

	log.Printf("[WARNING] Got %d new environments to be added", len(newEnvironments))

	if len(newEnvironments) < 1 {
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "One environment is required"}`)))
		return
	}

	if project.Environment == "cloud" {
		//foundOrg, err := GetOrg(ctx, user.ActiveOrg.Id)
		//if err != nil {
		//	resp.WriteHeader(401)
		//	resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed find your organization"}`)))
		//	return
		//}

		// FIXME: Removed need for syncfeatures to be enabled
		// September 2022
		//_ = foundOrg

		//if !foundOrg.SyncFeatures.MultiEnv.Active {
		//	resp.WriteHeader(401)
		//	resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Adding multiple environments requires an active hybrid, enterprise or MSSP subscription"}`)))
		//	return
		//}
	}

	// Validate input here
	defaults := 0
	parsedEnvs := []Environment{}
	for _, env := range newEnvironments {
		if project.Environment == "cloud" && env.Type == "cloud" && env.Archived {
			log.Printf("[WARNING] User %s (%s) tried to disable the cloud environment", user.Username, user.Id)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Can't disable cloud environments"}`))
			return
		}

		if env.Default && env.Archived {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Can't disable default environment"}`))
			return
		}

		//if project.Environment == "cloud" && env.Type != "cloud" && len(env.Name) < 10 {
		//	log.Printf("[ERROR] Skipping env %s because length is shorter than 10", env.Name)
		//	continue
		//}

		if defaults > 0 {
			env.Default = false
		}

		if env.Default {
			defaults += 1
		}

		parsedEnvs = append(parsedEnvs, env)
	}

	newEnvironments = parsedEnvs

	openEnvironments := 0
	for _, item := range newEnvironments {
		if !item.Archived {
			openEnvironments += 1
		}
	}

	if openEnvironments < 1 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Can't archive all environments. Not deleting."}`))
		return
	}

	// Clear old data? Removed for archiving purpose. No straight deletion
	log.Printf("[INFO] Deleting %d original environments before resetting. To be added: %d!", len(environments), len(newEnvironments))
	nameKey := "Environments"
	for _, item := range environments {
		DeleteKey(ctx, nameKey, item.Id)
		DeleteKey(ctx, nameKey, item.Name)
	}

	for _, item := range newEnvironments {
		for _, subenv := range environments {
			if item.Name == subenv.Name || item.Id == subenv.Id {
				item.Auth = subenv.Auth
				break
			}
		}

		item.RunningIp = ""
		if item.OrgId != user.ActiveOrg.Id {
			item.OrgId = user.ActiveOrg.Id
		}

		if len(item.Id) == 0 {
			item.Id = uuid.NewV4().String()
		}

		if len(item.Auth) == 0 {
			item.Auth = uuid.NewV4().String()
		}

		err = SetEnvironment(ctx, &item)
		if err != nil {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed setting environment variable"}`))
			return
		}
	}

	cacheKey := fmt.Sprintf("Environments_%s", user.ActiveOrg.Id)
	DeleteCache(ctx, cacheKey)

	log.Printf("[INFO] Set %d new environments for org %s", len(newEnvironments), user.ActiveOrg.Id)

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

func HandleRerunExecutions(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in stop executions: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if user.Role != "admin" {
		log.Printf("[AUDIT] User isn't admin during stop executions")
		resp.WriteHeader(409)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Must be admin to perform this action"}`)))
		return
	}

	if strings.ToLower(os.Getenv("SHUFFLE_DISABLE_RERUN_AND_ABORT")) == "true" {
		log.Printf("[AUDIT] Rerunning is disabled by the SHUFFLE_DISABLE_RERUN_AND_ABORT argument. Stopping.")
		resp.WriteHeader(409)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "SHUFFLE_DISABLE_RERUN_AND_ABORT is active. Won't rerun executions."}`)))
		return
	}

	ctx := GetContext(request)
	environmentName := fileId
	if len(fileId) != 36 {
		log.Printf("[DEBUG] Environment length %d for %s is not good for reruns. Attempting to find the actual ID for it", len(fileId), fileId)

		environments, err := GetEnvironments(ctx, user.ActiveOrg.Id)
		if err != nil {
			log.Printf("[WARNING] Failed getting environments to validate: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed to validate environment"}`))
			return
		}

		for _, environment := range environments {
			if environment.Name == fileId && len(environment.Id) > 0 {
				environmentName = fileId
				fileId = environment.Id
				break
			}
		}

		if len(fileId) != 36 {
			log.Printf("[WARNING] Failed getting environments to validate. New FileId: %s", fileId)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed updating environment"}`))
			return
		}
	}

	// 1: Loop all workflows
	// 2: Stop all running executions (manually abort)
	workflows, err := GetAllWorkflowsByQuery(ctx, user)
	if err != nil {
		log.Printf("[WARNING] Failed getting workflows for user %s (0): %s", user.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	maxTotalReruns := 100
	total := 0
	for _, workflow := range workflows {
		if workflow.OrgId != user.ActiveOrg.Id {
			//log.Printf("[DEBUG] Skipping workflow for org %s (user: %s)", workflow.OrgId, user.Username)
			continue
		}

		if total > maxTotalReruns {
			log.Printf("[DEBUG] Stopping because more than %d (%d) executions are pending. Checking reruns again on next iteration", maxTotalReruns, total)
			break
		}

		cnt, _ := RerunExecution(ctx, environmentName, workflow)
		total += cnt
	}

	//log.Printf("[DEBUG] RERAN %d execution(s) in total for environment %s for org %s", total, fileId, user.ActiveOrg.Id)
	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true, "reason": "Successfully RERAN %d executions"}`, total)))
}

// Send in deleteall=true to delete ALL executions for the environment ID
func HandleStopExecutions(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in stop executions: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
		if strings.Contains(fileId, "?") {
			fileId = strings.Split(fileId, "?")[0]
		}
	}

	if user.Role != "admin" {
		log.Printf("[AUDIT] User isn't admin during stop executions")
		resp.WriteHeader(409)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Must be admin to perform this action"}`)))
		return
	}

	// Fix here by allowing cleanup from UI anyway :)
	if strings.ToLower(os.Getenv("SHUFFLE_DISABLE_RERUN_AND_ABORT")) == "true" {
		log.Printf("[AUDIT] Rerunning is disabled by the SHUFFLE_DISABLE_RERUN_AND_ABORT argument. Stopping. (abort)")
		resp.WriteHeader(409)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "SHUFFLE_DISABLE_RERUN_AND_ABORT is active. Won't rerun executions (abort)"}`)))
		return
	}

	ctx := GetContext(request)
	environmentName := fileId
	if len(fileId) != 36 {
		log.Printf("[DEBUG] Environment length %d for %s is not good for executions aborts. Attempting to find the actual ID for it", len(fileId), fileId)

		environments, err := GetEnvironments(ctx, user.ActiveOrg.Id)
		if err != nil {
			log.Printf("[WARNING] Failed getting environments to validate: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed to validate environment"}`))
			return
		}

		for _, environment := range environments {
			if environment.Name == fileId && len(environment.Id) > 0 {
				environmentName = fileId
				fileId = environment.Id
				break
			}
		}

		if len(fileId) != 36 {
			log.Printf("[WARNING] Failed getting environments to validate. New FileId: %s", fileId)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed updating environment"}`))
			return
		}
	}

	cleanAll := false
	deleteAll, ok := request.URL.Query()["deleteall"]
	if ok {
		if deleteAll[0] == "true" {
			cleanAll = true

			if project.Environment != "cloud" {
				log.Printf("[DEBUG] Deleting and aborting ALL executions for this environment and org %s!", user.ActiveOrg.Id)

				env, err := GetEnvironment(ctx, fileId, user.ActiveOrg.Id)
				if err != nil {
					log.Printf("[WARNING] Failed to get environment %s for org %s", fileId, user.ActiveOrg.Id)
					resp.WriteHeader(401)
					resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed to get environment %s"}`, fileId)))
					return
				}

				if env.OrgId != user.ActiveOrg.Id {
					log.Printf("[WARNING] %s (%s) doesn't have permission to stop all executions for environment %s", user.Username, user.Id, fileId)
					resp.WriteHeader(401)
					resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "You don't have permission to stop environment executions for ID %s"}`, fileId)))
					return
				}

				// If here, it should DEFINITELY clean up all executions
				// Runs on 10.000 workflows max
				maxAmount := 1000
				for i := 0; i < 10; i++ {
					executionRequests, err := GetWorkflowQueue(ctx, env.Name, maxAmount)
					if err != nil {
						log.Printf("[WARNING] Jumping out of workflowqueue delete handler: %s", err)
						break
					}

					if len(executionRequests.Data) == 0 {
						break
					}

					ids := []string{}
					for _, execution := range executionRequests.Data {
						if !ArrayContains(execution.Environments, env.Name) {
							continue
						}

						ids = append(ids, execution.ExecutionId)
					}

					parsedId := fmt.Sprintf("workflowqueue-%s", strings.ToLower(env.Name))
					err = DeleteKeys(ctx, parsedId, ids)
					if err != nil {
						log.Printf("[ERROR] Failed deleting %d execution keys for org %s during force stop: %s", len(ids), env.Name, err)
					} else {
						log.Printf("[INFO] Deleted %d keys from org %s during force stop", len(ids), parsedId)
					}

					if len(executionRequests.Data) != maxAmount {
						log.Printf("[DEBUG] Less than 1000 in queue. Not querying more")
						break
					}
				}
			}
		}
	}

	// 1: Loop all workflows
	// 2: Stop all running executions (manually abort)
	workflows, err := GetAllWorkflowsByQuery(ctx, user)
	if err != nil {
		log.Printf("[WARNING] Failed getting workflows for user %s (0): %s", user.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	total := 0
	for _, workflow := range workflows {
		if workflow.OrgId != user.ActiveOrg.Id {
			log.Printf("[DEBUG] Skipping workflow for org %s (user: %s)", workflow.OrgId, user.Username)
			continue
		}

		cnt, _ := CleanupExecutions(ctx, environmentName, workflow, cleanAll)
		total += cnt
	}

	if total > 0 {
		log.Printf("[DEBUG] Stopped %d executions in total for environment %s for org %s", total, fileId, user.ActiveOrg.Id)
	}

	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true, "reason": "Successfully deleted and stopped %d executions"}`, total)))
}

func RerunExecution(ctx context.Context, environment string, workflow Workflow) (int, error) {
	maxReruns := 100
	//log.Printf("[DEBUG] Finding executions for %s", workflow.ID)
	//if project.Environment == "cloud" && workflow.ID != "0a0752db-61ba-49ea-a488-4dfe6d2f57a0" {
	//	log.Printf("Bad workflow ID during test: %s", workflow.ID)
	//	return 0, nil
	//}

	executions, err := GetUnfinishedExecutions(ctx, workflow.ID)
	if err != nil {
		log.Printf("[DEBUG] Failed getting executions for workflow %s", workflow.ID)
		return 0, err
	}

	if len(executions) == 0 {
		return 0, nil
	}

	//log.Printf("[DEBUG] Found %d POTENTIALLY unfinished executions for workflow %s (%s) with environment %s that are more than 30 minutes old", len(executions), workflow.Name, workflow.ID, environment)
	//log.Printf("[DEBUG] Found %d unfinished executions for workflow %s (%s) with environment %s that are more than 30 minutes old", len(executions), workflow.Name, workflow.ID, environment)

	//backendUrl := os.Getenv("BASE_URL")
	//if project.Environment == "cloud" {
	//	backendUrl = "https://shuffler.io"
	//} else {
	//	backendUrl = "http://127.0.0.1:5001"
	//}

	//topClient := &http.Client{
	//	Transport: &http.Transport{
	//		Proxy: nil,
	//	},
	//}
	//_ = backendUrl
	//_ = topClient

	//StartedAt           int64          `json:"started_at" datastore:"started_at"`
	timeNow := int64(time.Now().Unix())
	cnt := 0

	// Rerun after 570 seconds (9.5 minutes), ensuring it can check 3 times before
	// automated aborting of the execution happens
	waitTime := 270
	//waitTime := 0
	executed := []string{}
	for _, execution := range executions {
		if timeNow < execution.StartedAt+int64(waitTime) {
			//log.Printf("Bad timing: %d", execution.StartedAt)
			continue
		}

		if execution.Status != "EXECUTING" {
			//log.Printf("Bad status: %s", execution.Status)
			continue
		}

		if ArrayContains(executed, execution.ExecutionId) {
			continue
		}

		executed = append(executed, execution.ExecutionId)

		found := false
		environments := []string{}
		for _, action := range execution.Workflow.Actions {
			if action.Environment == environment {
				environments = append(environments, action.Environment)
				found = true
				break
			}
		}

		if len(environments) == 0 {
			found = true
		}

		if !found {
			continue
		}

		if cnt > maxReruns {
			log.Printf("[DEBUG] Breaking because more than 100 executions are executing")
			break
		}

		if project.Environment != "cloud" {
			executionRequest := ExecutionRequest{
				ExecutionId:   execution.ExecutionId,
				WorkflowId:    execution.Workflow.ID,
				Authorization: execution.Authorization,
				Environments:  environments,
			}

			executionRequest.Priority = execution.Priority
			err = SetWorkflowQueue(ctx, executionRequest, environment)
			if err != nil {
				log.Printf("[ERROR] Failed re-adding execution to db: %s", err)
			}
		} else {
			//log.Printf("[DEBUG] Rerunning executions is not available in cloud yet.")
			//if len(environments) != 1 || strings.ToLower(environments[0]) != "cloud" {
			//	log.Printf("[DEBUG][%s] Skipping execution for workflow %s because it's not for JUST the cloud env. Org: %s", execution.ExecutionId, execution.Workflow.ID, execution.OrgId)
			//	continue
			//}

			streamUrl := fmt.Sprintf("https://shuffler.io")
			if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
				streamUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
			}

			if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
				streamUrl = fmt.Sprintf("%s", os.Getenv("SHUFFLE_CLOUDRUN_URL"))
			}

			streamUrl = fmt.Sprintf("%s/api/v1/workflows/%s/executions/%s/rerun", streamUrl, execution.Workflow.ID, execution.ExecutionId)

			client := &http.Client{
				Timeout: 5 * time.Second,
			}
			req, err := http.NewRequest(
				"POST",
				streamUrl,
				nil,
			)

			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", execution.Authorization))
			if err != nil {
				log.Printf("[WARNING] Error in new request for manual rerun: %s", err)
				continue
			}

			newresp, err := client.Do(req)
			if err != nil {
				log.Printf("[WARNING] Error running body for manual rerun: %s", err)
				continue
			}

			body, err := ioutil.ReadAll(newresp.Body)
			if err != nil {
				log.Printf("[WARNING] Failed reading body for manual rerun: %s", err)
				continue
			}

			log.Printf("[DEBUG] Rerun response: %s", string(body))
		}

		cnt += 1
		log.Printf("[DEBUG] Should rerun execution %s (%s - Workflow: %s) with environments %s", execution.ExecutionId, execution.Status, execution.Workflow.ID, environments)
		//log.Printf("[DEBUG] Result from rerunning %s: %s", execution.ExecutionId, string(body))
	}

	return cnt, nil
}

func CleanupExecutions(ctx context.Context, environment string, workflow Workflow, cleanAll bool) (int, error) {
	executions, err := GetUnfinishedExecutions(ctx, workflow.ID)
	if err != nil {
		log.Printf("[DEBUG] Failed getting executions for workflow %s", workflow.ID)
		return 0, err
	}

	if len(executions) == 0 {
		return 0, nil
	}

	//log.Printf("[DEBUG] Found %d POTENTIALLY unfinished executions for workflow %s (%s) with environment %s that are more than 30 minutes old", len(executions), workflow.Name, workflow.ID, environment)
	//log.Printf("[DEBUG] Found %d unfinished executions for workflow %s (%s) with environment %s that are more than 30 minutes old", len(executions), workflow.Name, workflow.ID, environment)

	backendUrl := os.Getenv("BASE_URL")
	// Redundant, but working ;)
	if project.Environment == "cloud" {
		backendUrl = "https://shuffler.io"

		if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
			backendUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
		}

		if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
			backendUrl = os.Getenv("SHUFFLE_CLOUDRUN_URL")
		}

	} else {
		backendUrl = "http://127.0.0.1:5001"
	}

	topClient := &http.Client{
		Transport: &http.Transport{
			Proxy: nil,
		},
	}

	//StartedAt           int64          `json:"started_at" datastore:"started_at"`
	timeNow := int64(time.Now().Unix())
	cnt := 0
	for _, execution := range executions {
		if cleanAll {
		} else if timeNow < execution.StartedAt+1800 {
			//log.Printf("Bad timing: %d", execution.StartedAt)
			continue
		}

		if execution.Status != "EXECUTING" {
			//log.Printf("Bad status: %s", execution.Status)
			continue
		}

		found := false
		environments := []string{}
		for _, action := range execution.Workflow.Actions {
			if action.Environment == environment {
				environments = append(environments, action.Environment)
				found = true
				break
			}
		}

		if len(environments) == 0 {
			found = true
		}

		if !found {
			continue
		}

		//log.Printf("[DEBUG] Got execution with status %s!", execution.Status)

		streamUrl := fmt.Sprintf("%s/api/v1/workflows/%s/executions/%s/abort?reason=%s", backendUrl, execution.Workflow.ID, execution.ExecutionId, url.QueryEscape(`{"success": False, "reason": "Shuffle's automated cleanup bot stopped this execution as it didn't finish within 30 minutes."}`))
		//log.Printf("Url: %s", streamUrl)
		req, err := http.NewRequest(
			"GET",
			streamUrl,
			nil,
		)

		if err != nil {
			log.Printf("[ERROR] Error in auto-abort request: %s", err)
			continue
		}

		req.Header.Add("Authorization", fmt.Sprintf(`Bearer %s`, execution.Authorization))
		newresp, err := topClient.Do(req)
		if err != nil {
			log.Printf("[ERROR] Error auto-aborting workflow: %s", err)
			continue
		}

		body, err := ioutil.ReadAll(newresp.Body)
		if err != nil {
			log.Printf("[ERROR] Failed reading parent body: %s", err)
			continue
		}
		//log.Printf("BODY (%d): %s", newresp.StatusCode, string(body))

		if newresp.StatusCode != 200 {
			log.Printf("[ERROR] Bad statuscode in auto-abort: %d, %s", newresp.StatusCode, string(body))
			continue
		}

		cnt += 1
		log.Printf("[DEBUG] Result from aborting %s: %s", execution.ExecutionId, string(body))
	}

	return cnt, nil
}

func HandleGetEnvironments(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get environments: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)
	environments, err := GetEnvironments(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Failed getting environments")
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Can't get environments"}`))
		return
	}

	// Always make Cloud the default environment
	// If there are multiple
	if project.Environment == "cloud" {
		defaults := []int{}
		cloudFound := false
		for envIndex, environment := range environments {
			if environment.Default {
				defaults = append(defaults, envIndex)
			}

			if strings.ToLower(environment.Name) == "cloud" {
				cloudFound = true
			}
		}

		// Ensure it's attached. When they click "set as default", it will become activated forever :>
		// Found by seeing a user from early on that didn't have the env
		if !cloudFound {
			setDefault := false
			if len(environments) == 1 || len(defaults) == 0 {
				setDefault = true
			}

			environments = append(environments, Environment{
				Name:       "Cloud",
				Type:       "cloud",
				Archived:   false,
				Registered: true,
				Default:    setDefault,
				OrgId:      user.ActiveOrg.Id,
				Id:         uuid.NewV4().String(),
			})

			defaults = append(defaults, len(environments)-1)
		}

		// Fallback to cloud for now
		if len(defaults) > 1 {
			for _, index := range defaults {
				if strings.ToLower(environments[index].Name) == "cloud" {
					continue
				} else {
					environments[index].Default = false
				}
			}
		}
	}

	newEnvironments := []Environment{}
	for _, environment := range environments {
		if len(environment.Id) == 0 {
			environment.Id = uuid.NewV4().String()
		}

		found := false
		for _, oldEnv := range newEnvironments {
			if oldEnv.Name == environment.Name {
				found = true
			}
		}

		if !found {
			newEnvironments = append(newEnvironments, environment)
		}
	}

	newjson, err := json.Marshal(newEnvironments)
	if err != nil {
		log.Printf("Failed unmarshal: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking environments"}`)))
		return
	}

	//log.Printf("Existing environments: %s", string(newjson))

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func HandleApiAuthentication(resp http.ResponseWriter, request *http.Request) (User, error) {
	apikey := request.Header.Get("Authorization")

	ctx := GetContext(request)
	user := &User{}
	if len(apikey) > 0 {
		if !strings.HasPrefix(apikey, "Bearer ") {

			//location := strings.Split(request.URL.String(), "/")
			if !strings.Contains(request.URL.String(), "/execute") {
				log.Printf("[WARNING] Apikey doesn't start with bearer")
			}

			return User{}, errors.New("No bearer token for authorization header")
		}

		apikeyCheck := strings.Split(apikey, " ")
		if len(apikeyCheck) != 2 {
			log.Printf("[WARNING] Invalid format for apikey: %s", apikeyCheck)
			return User{}, errors.New("Invalid format for apikey")
		}

		if len(apikeyCheck[1]) < 36 {
			return User{}, errors.New("Apikey must be at least 36 characters long (UUID)")
		}

		// This is annoying af
		newApikey := apikeyCheck[1]
		if len(newApikey) > 249 {
			newApikey = newApikey[0:248]
		}

		cache, err := GetCache(ctx, newApikey)
		if err == nil {
			cacheData := []byte(cache.([]uint8))
			err = json.Unmarshal(cacheData, &user)
			if err == nil {
				//log.Printf("[WARNING] Got user from cache: %s", err)

				if len(user.Id) == 0 && len(user.Username) == 0 {
					return User{}, errors.New(fmt.Sprintf("Couldn't find user"))
				}

				user.ApiKey = newApikey
				user.SessionLogin = false
				return *user, nil
			}
		} else {
			//log.Printf("[WARNING] Error getting authentication cache for %s: %v", newApikey, err)
		}

		// Make specific check for just service user?
		// Get the user based on APIkey here
		userdata, err := GetApikey(ctx, apikeyCheck[1])
		if err != nil {
			//log.Printf("[WARNING] Apikey %s doesn't exist: %s", apikey, err)
			return User{}, err
		}

		if len(userdata.Id) == 0 && len(userdata.Username) == 0 {
			//log.Printf("[WARNING] Apikey %s doesn't exist or the user doesn't have an ID/Username", apikey)
			return User{}, errors.New("Couldn't find the user")
		}

		// Caching both bad and good apikeys :)
		b, err := json.Marshal(userdata)
		if err != nil {
			log.Printf("[WARNING] Failed marshalling: %s", err)
			return User{}, err
		}

		user.SessionLogin = false
		user.ApiKey = newApikey
		err = SetCache(ctx, newApikey, b, 30)
		if err != nil {
			log.Printf("[WARNING] Failed setting cache for apikey: %s", err)
		}

		return userdata, nil
	}

	// One time API keys
	//authorizationArr, ok := request.URL.Query()["authorization"]
	//if ok {
	//	//authorization := ""
	//	//if len(authorizationArr) > 0 {
	//	//	authorization = authorizationArr[0]
	//	//}
	//	//_ = authorization
	//	//log.Printf("[ERROR] WHAT ARE ONE TIME KEYS USED FOR? User input?")
	//}

	// __session first due to Compatibility issues
	c, err := request.Cookie("__session")
	if err != nil {
		c, err = request.Cookie("session_token")
	}

	if err == nil {
		sessionToken := c.Value

		newCookie := &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(-100 * time.Hour),
			MaxAge:  -1,
		}

		if project.Environment == "cloud" {
			newCookie.Domain = ".shuffler.io"
			newCookie.Secure = true
			newCookie.HttpOnly = true
		}

		user, err := GetSessionNew(ctx, sessionToken)
		if err != nil {
			log.Printf("[ERROR] No valid session token for ID %s. Setting cookie to expire. May cause fallback problems.", sessionToken)

			if resp != nil {
				http.SetCookie(resp, newCookie)

				newCookie.Name = "__session"
				http.SetCookie(resp, newCookie)
			}

			return User{}, err
		} else {
			// Check if both session tokens are set
			// Compatibility issues
			//expiration := time.Now().Add(3600 * time.Second)
			newCookie.Expires = c.Expires
			newCookie.MaxAge = c.MaxAge

			_, err1 := request.Cookie("session_token")
			if err1 != nil {
				log.Printf("[DEBUG] Setting missing session_token for user %s (%s) (1)", user.Username, user.Id)
				newCookie.Name = "session_token"
				if resp != nil {
					http.SetCookie(resp, newCookie)
				}
			}

			_, err2 := request.Cookie("__session")
			if err2 != nil {
				log.Printf("[DEBUG] Setting missing __session for user %s (%s) (2)", user.Username, user.Id)
				newCookie.Name = "__session"
				if resp != nil {
					http.SetCookie(resp, newCookie)
				}
			}
		}

		if len(user.Id) == 0 && len(user.Username) == 0 {

			newCookie := &http.Cookie{
				Name:    "session_token",
				Value:   sessionToken,
				Expires: time.Now().Add(-100 * time.Hour),
				MaxAge:  -1,
			}

			if project.Environment == "cloud" {
				newCookie.Domain = ".shuffler.io"
				newCookie.Secure = true
				newCookie.HttpOnly = true
			}

			if resp != nil {
				http.SetCookie(resp, newCookie)

				newCookie.Name = "__session"
				http.SetCookie(resp, newCookie)
			}

			return User{}, errors.New(fmt.Sprintf("Couldn't find user"))
		}

		// We're using the session to find the user anyway, which is NOT user controlled
		// This means that this is redundant, but MAY allow users
		// to have access past session timeouts
		//if user.Session != sessionToken {
		//	return User{}, errors.New("[WARNING] Wrong session token")
		//}

		user.SessionLogin = true

		// Means session exists, but
		return user, nil
	}

	// Key = apikey
	return User{}, errors.New("Missing authentication")
}

func GetOpenapi(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Just here to verify that the user is logged in
	_, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in validate swagger: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")
	var id string
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Missing parts of API in request!")
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		id = location[4]
	}

	/*
		if len(id) != 32 {
			log.Printf("Missing parts of API in request!")
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	*/
	//_, err = GetApp(ctx, id)
	//if err == nil {
	//}

	// FIXME - FIX AUTH WITH APP
	ctx := GetContext(request)
	parsedApi, err := GetOpenApiDatastore(ctx, id)
	if err != nil {
		log.Printf("[ERROR] Failed getting OpenAPI: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	log.Printf("[INFO] API LENGTH GET: %d, ID: %s", len(parsedApi.Body), id)

	parsedApi.Success = true
	data, err := json.Marshal(parsedApi)
	if err != nil {
		log.Printf("[ERROR] Failed unmarshaling OpenAPI: %s", err)
		resp.WriteHeader(422)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed marshalling parsed swagger: %s"}`, err)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(data)
}

func GetResult(ctx context.Context, workflowExecution WorkflowExecution, id string) (WorkflowExecution, ActionResult) {
	for _, actionResult := range workflowExecution.Results {
		if actionResult.Action.ID == id {
			// ALWAYS relying on cache due to looping subflow issues
			if actionResult.Status == "WAITING" && actionResult.Action.AppName == "User Input" {
				break
			}

			if actionResult.Action.AppName == "shuffle-subflow" && project.Environment == "cloud" {
				//if os.Getenv("SHUFFLE_SWARM_CONFIG") == "run" && (project.Environment == "" || project.Environment == "worker") {
				//log.Printf("[INFO] Skipping due to cache requirement for subflow")
				break
			}

			return workflowExecution, actionResult
		}
	}

	//log.Printf("[WARNING] No result found for %s - add here too?", id)
	cacheId := fmt.Sprintf("%s_%s_result", workflowExecution.ExecutionId, id)
	cache, err := GetCache(ctx, cacheId)
	if err == nil {
		//log.Printf("[DEBUG] Already found %s executed, but not in list.. Adding!\n\n\n\n\n", id)

		actionResult := ActionResult{}
		cacheData := []byte(cache.([]uint8))
		// Just ensuring the data is good
		err = json.Unmarshal(cacheData, &actionResult)
		if err == nil {
			workflowExecution.Results = append(workflowExecution.Results, actionResult)
			SetWorkflowExecution(ctx, workflowExecution, false)
			return workflowExecution, actionResult
		}
	}

	return workflowExecution, ActionResult{}
}

func GetAction(workflowExecution WorkflowExecution, id, environment string) Action {
	for _, action := range workflowExecution.Workflow.Actions {
		if action.ID == id {
			return action
		}
	}

	for _, trigger := range workflowExecution.Workflow.Triggers {
		if trigger.ID == id {
			return Action{
				ID:          trigger.ID,
				AppName:     trigger.AppName,
				Name:        trigger.AppName,
				Environment: environment,
				Label:       trigger.Label,
			}
			log.Printf("FOUND TRIGGER: %s!", trigger)
		}
	}

	return Action{}
}

func ArrayContainsLower(visited []string, id string) bool {
	found := false
	for _, item := range visited {
		if strings.ToLower(item) == strings.ToLower(id) {
			found = true
		}
	}

	return found
}

func ArrayContains(visited []string, id string) bool {
	found := false
	for _, item := range visited {
		if item == id {
			found = true
		}
	}

	return found
}

func GetWorkflowExecutions(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in getting workflow executions: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")

	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if len(fileId) != 36 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow ID when getting workflow executions is not valid"}`))
		return
	}

	ctx := GetContext(request)
	workflow, err := GetWorkflow(ctx, fileId)
	if err != nil {
		log.Printf("[WARNING] Failed getting the workflow %s locally (get executions): %s", fileId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Id != workflow.Owner || len(user.Id) == 0 {
		if workflow.OrgId == user.ActiveOrg.Id {
			log.Printf("[AUDIT] User %s is accessing workflow '%s' (%s) executions as %s (get executions)", user.Username, workflow.Name, workflow.ID, user.Role)
		} else if project.Environment == "cloud" && user.Verified == true && user.Active == true && user.SupportAccess == true && strings.HasSuffix(user.Username, "@shuffler.io") {
			log.Printf("[AUDIT] Letting verified support admin %s access workflow execs for %s", user.Username, workflow.ID)
		} else {
			log.Printf("[AUDIT] Wrong user (%s) for workflow %s (get workflow execs)", user.Username, workflow.ID)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	// Query for the specifci workflowId
	//q := datastore.NewQuery("workflowexecution").Filter("workflow_id =", fileId).Order("-started_at").Limit(30)
	//q := datastore.NewQuery("workflowexecution").Filter("workflow_id =", fileId)
	maxAmount := 100
	top, topOk := request.URL.Query()["top"]
	if topOk && len(top) > 0 {
		val, err := strconv.Atoi(top[0])
		if err == nil {
			maxAmount = val
		}
	}

	if maxAmount > 1000 {
		maxAmount = 1000
	}

	workflowExecutions, err := GetAllWorkflowExecutions(ctx, fileId, maxAmount)
	if err != nil {
		log.Printf("[WARNING] Failed getting executions for %s", fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	//log.Printf("[DEBUG] Found %d executions for workflow %s", len(workflowExecutions), fileId)

	if len(workflowExecutions) == 0 {
		resp.WriteHeader(200)
		resp.Write([]byte("[]"))
		return
	}

	for index, execution := range workflowExecutions {
		newResults := []ActionResult{}
		newActions := []Action{}

		// Results
		for _, result := range execution.Results {
			newParams := []WorkflowAppActionParameter{}
			for _, param := range result.Action.Parameters {
				if param.Configuration || strings.Contains(strings.ToLower(param.Name), "user") || strings.Contains(strings.ToLower(param.Name), "key") || strings.Contains(strings.ToLower(param.Name), "pass") {
					param.Value = ""
					//log.Printf("FOUND CONFIG: %s!!", param.Name)
				}

				newParams = append(newParams, param)
			}

			result.Action.Parameters = newParams
			newResults = append(newResults, result)
		}

		// Actions
		for _, action := range execution.Workflow.Actions {
			newParams := []WorkflowAppActionParameter{}
			for _, param := range action.Parameters {
				if param.Configuration || strings.Contains(strings.ToLower(param.Name), "user") || strings.Contains(strings.ToLower(param.Name), "key") || strings.Contains(strings.ToLower(param.Name), "pass") {
					param.Value = ""
					//log.Printf("FOUND CONFIG: %s!!", param.Name)
				}

				newParams = append(newParams, param)
			}

			action.Parameters = newParams
			newActions = append(newActions, action)
		}

		workflowExecutions[index].Results = newResults
		workflowExecutions[index].Workflow.Actions = newActions
	}

	newjson, err := json.Marshal(workflowExecutions)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking workflow executions"}`)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func GetWorkflows(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in getworkflows: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)
	var workflows []Workflow

	cacheKey := fmt.Sprintf("%s_workflows", user.Id)
	cache, err := GetCache(ctx, cacheKey)
	if err == nil {
		cacheData := []byte(cache.([]uint8))
		err = json.Unmarshal(cacheData, &workflows)
		if err == nil {
			resp.WriteHeader(200)
			resp.Write(cacheData)
			return
		}
	} else {
		//log.Printf("[INFO] Failed getting cache for workflows for user %s", user.Id)
	}

	workflows, err = GetAllWorkflowsByQuery(ctx, user)
	if err != nil {
		log.Printf("[WARNING] Failed getting workflows for user %s (0): %s", user.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if len(workflows) == 0 {
		log.Printf("[INFO] No workflows found for user %s", user.Username)
		resp.WriteHeader(200)
		resp.Write([]byte("[]"))
		return
	}

	usecaseIds := []string{}
	newWorkflows := []Workflow{}
	for _, workflow := range workflows {
		if workflow.OrgId != user.ActiveOrg.Id {
			continue
		}

		newActions := []Action{}
		for _, action := range workflow.Actions {
			// Removed because of exports. These are needed there.
			//action.LargeImage = ""
			//action.SmallImage = ""
			action.ReferenceUrl = ""
			newActions = append(newActions, action)
		}

		workflow.Actions = newActions

		// Skipping these as they're related to onprem workflows in cloud (orborus)
		if project.Environment == "cloud" && workflow.ExecutionEnvironment == "onprem" {
			continue
		}

		usecaseIds = append(usecaseIds, workflow.UsecaseIds...)
		newWorkflows = append(newWorkflows, workflow)
	}

	// Get the org as well to manage priorities
	// Only happens on first load, so it's like once per session~
	if len(usecaseIds) > 0 {
		org, err := GetOrg(ctx, user.ActiveOrg.Id)
		if err != nil {
			log.Printf("[WARNING] Failed getting org %s for user %s during workflow load: %s", user.ActiveOrg.Id, user.Username, err)
		} else {
			for prioIndex, priority := range org.Priorities {
				if priority.Type != "usecase" || priority.Active != true {
					continue
				}

				for _, usecaseId := range usecaseIds {
					if strings.Contains(strings.ToLower(priority.Name), strings.ToLower(usecaseId)) {
						//log.Printf("\n\n[DEBUG] Found usecase %s in priority %s\n\n", usecaseId, priority.Name)
						org.Priorities[prioIndex].Active = false

						SetOrg(ctx, *org, org.Id)
						break
					}
				}
			}
		}
	}

	//log.Printf("[INFO] Returning %d workflows", len(newWorkflows))
	newjson, err := json.Marshal(newWorkflows)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking workflows"}`)))
		return
	}

	if project.CacheDb {
		err = SetCache(ctx, cacheKey, newjson, 30)
		if err != nil {
			log.Printf("[WARNING] Failed updating workflow cache: %s", err)
		}
	}

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func SetAuthenticationConfig(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, userErr := HandleApiAuthentication(resp, request)
	if userErr != nil {
		log.Printf("[AUDIT] Api authentication failed in get all apps: %s", userErr)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role != "admin" {
		log.Printf("[AUDIT] User isn't admin during auth edit config")
		resp.WriteHeader(409)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Must be admin to perform this action"}`)))
		return
	}

	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 5 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[5]
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error with body read: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	type configAuth struct {
		Id     string `json:"id"`
		Action string `json:"action"`
	}

	var config configAuth
	err = json.Unmarshal(body, &config)
	if err != nil {
		log.Printf("Failed unmarshaling (appauth): %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if config.Id != fileId {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Bad ID match"}`))
		return
	}

	ctx := GetContext(request)
	auth, err := GetWorkflowAppAuthDatastore(ctx, fileId)
	if err != nil {
		log.Printf("Authget error: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": ":("}`))
		return
	}

	if auth.OrgId != user.ActiveOrg.Id {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "User can't edit this org"}`))
		return
	}

	if config.Action == "assign_everywhere" {
		log.Printf("[INFO] Should set authentication config")
		baseWorkflows, err := GetAllWorkflowsByQuery(ctx, user)
		if err != nil && len(baseWorkflows) == 0 {
			log.Printf("Getall error in auth update: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed getting workflows to update"}`))
			return
		}

		workflows := []Workflow{}
		for _, workflow := range baseWorkflows {
			if workflow.OrgId == user.ActiveOrg.Id {
				workflows = append(workflows, workflow)
			}
		}

		// FIXME: Add function to remove auth from other auth's
		actionCnt := 0
		workflowCnt := 0
		authenticationUsage := []AuthenticationUsage{}
		for _, workflow := range workflows {
			newActions := []Action{}
			edited := false
			usage := AuthenticationUsage{
				WorkflowId: workflow.ID,
				Nodes:      []string{},
			}

			for _, action := range workflow.Actions {
				if action.AppName == auth.App.Name {
					action.AuthenticationId = auth.Id

					edited = true
					actionCnt += 1
					usage.Nodes = append(usage.Nodes, action.ID)
				}

				newActions = append(newActions, action)
			}

			workflow.Actions = newActions
			if edited {
				//auth.Usage = usage
				authenticationUsage = append(authenticationUsage, usage)
				err = SetWorkflow(ctx, workflow, workflow.ID)
				if err != nil {
					log.Printf("Failed setting (authupdate) workflow: %s", err)
					continue
				}

				cacheKey := fmt.Sprintf("%s_workflows", user.Id)

				DeleteCache(ctx, cacheKey)

				workflowCnt += 1
			}
		}

		//Usage         []AuthenticationUsage `json:"usage" datastore:"usage"`
		log.Printf("[INFO] Found %d workflows, %d actions", workflowCnt, actionCnt)
		if actionCnt > 0 && workflowCnt > 0 {
			auth.WorkflowCount = int64(workflowCnt)
			auth.NodeCount = int64(actionCnt)
			auth.Usage = authenticationUsage
			auth.Defined = true

			err = SetWorkflowAppAuthDatastore(ctx, *auth, auth.Id)
			if err != nil {
				log.Printf("Failed setting appauth: %s", err)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false, "reason": "Failed setting app auth for all workflows"}`))
				return
			} else {
				// FIXME: Remove ALL workflows from other auths using the same
			}
		}
	}

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
	//var config configAuth

	//log.Printf("Should set %s
}

func HandleGetSchedules(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get schedules: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role != "admin" {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Admin required"}`))
		return
	}

	ctx := GetContext(request)
	schedules, err := GetAllSchedules(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Failed getting schedules: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Couldn't get schedules"}`))
		return
	}

	newjson, err := json.Marshal(schedules)
	if err != nil {
		log.Printf("Failed unmarshal: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking environments"}`)))
		return
	}

	//log.Printf("Existing environments: %s", string(newjson))

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func HandleUpdateUser(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains

		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting Update User request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	userInfo, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[AUDIT] Api authentication failed in update user: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Failed reading body in update user: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Required field: user_id"}`)))
		return
	}

	// NEVER allow the user to set all the data themselves
	type newUserStruct struct {
		Tutorial    string   `json:"tutorial" datastore:"tutorial"`
		Firstname   string   `json:"firstname"`
		Lastname    string   `json:"lastname"`
		Role        string   `json:"role"`
		Username    string   `json:"username"`
		UserId      string   `json:"user_id"`
		EthInfo     EthInfo  `json:"eth_info"`
		CompanyRole string   `json:"company_role"`
		Suborgs     []string `json:"suborgs"`

		CreatorDescription string          `json:"creator_description"`
		CreatorUrl         string          `json:"creator_url"`
		CreatorLocation    string          `json:"creator_location"`
		CreatorSkills      string          `json:"creator_skills"`
		CreatorWorkStatus  string          `json:"creator_work_status"`
		CreatorSocial      string          `json:"creator_social"`
		SpecializedApps    []MinimizedApps `json:"specialized_apps"`
	}

	ctx := GetContext(request)
	var t newUserStruct
	err = json.Unmarshal(body, &t)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshaling userId: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unmarshaling. Required field: user_id"}`)))
		return
	}

	// Should this role reflect the users' org access?
	// When you change org -> change user role
	if userInfo.Role != "admin" && userInfo.Id != t.UserId {
		log.Printf("[WARNING] User %s tried to update user %s. Role: %s", userInfo.Username, t.UserId, userInfo.Role)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "You need to be admin to change other users"}`)))
		return
	}

	foundUser, err := GetUser(ctx, t.UserId)
	if err != nil {
		log.Printf("[WARNING] Can't find user %s (update user): %s", t.UserId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false}`)))
		return
	}

	defaultRole := foundUser.Role
	orgFound := false
	for _, item := range foundUser.Orgs {
		if item == userInfo.ActiveOrg.Id {
			orgFound = true
			break
		}
	}

	if !orgFound {
		log.Printf("[AUDIT] User %s is admin, but can't edit users outside their own org.", userInfo.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can't change users outside your org."}`)))
		return
	}

	orgUpdater := true
	if len(t.Role) > 0 && (t.Role != "admin" && t.Role != "user" && t.Role != "org-reader") {
		log.Printf("[WARNING] %s tried and failed to update user %s", userInfo.Username, t.UserId)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can only change to roles user, admin and org-reader"}`)))
		return
	} else {
		// Same user - can't edit yourself?
		if len(t.Role) > 0 && (userInfo.Id == t.UserId || userInfo.Username == t.UserId) {
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can't update the role of your own user"}`)))
			return
		}

		if len(t.Role) > 0 {
			log.Printf("[INFO] Updated user %s from %s to %s in org %s. If role is empty, not updating", foundUser.Username, foundUser.Role, t.Role, userInfo.ActiveOrg.Id)
			orgUpdater = false

			// Realtime update if the user is in the same org
			if userInfo.ActiveOrg.Id == foundUser.ActiveOrg.Id {
				foundUser.Role = t.Role
				foundUser.Roles = []string{t.Role}
			}

			// Getting the specific org and just updating the user in that one
			foundOrg, err := GetOrg(ctx, userInfo.ActiveOrg.Id)
			if err != nil {
				log.Printf("[WARNING] Failed to get org in edit role to %s for %s (%s): %s", t.Role, foundUser.Username, foundUser.Id, err)
			} else {
				users := []User{}
				for _, user := range foundOrg.Users {
					if user.Id == foundUser.Id {
						user.Role = t.Role
						user.Roles = []string{t.Role}
					}

					users = append(users, user)
				}

				foundOrg.Users = users
				err = SetOrg(ctx, *foundOrg, foundOrg.Id)
				if err != nil {
					log.Printf("[WARNING] Failed setting org when changing role to %s for %s (%s): %s", t.Role, foundUser.Username, foundUser.Id, err)
				}
			}
		}
	}

	if len(t.Username) > 0 && project.Environment != "cloud" {
		users, err := FindUser(ctx, strings.ToLower(strings.TrimSpace(t.Username)))
		if err != nil && len(users) == 0 {
			log.Printf("[WARNING] Failed getting user %s: %s", t.Username, err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Username and/or password is incorrect"}`))
			return
		}

		found := false
		for _, item := range users {
			if item.Username == t.Username {
				found = true
				break
			}
		}

		if found {
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "User with username %s already exists"}`, t.Username)))
			return
		}

		foundUser.Username = t.Username
		if foundUser.Role == "" {
			foundUser.Role = defaultRole
		}
	}

	if len(t.Tutorial) > 0 {
		if !ArrayContains(foundUser.PersonalInfo.Tutorials, t.Tutorial) {
			foundUser.PersonalInfo.Tutorials = append(foundUser.PersonalInfo.Tutorials, t.Tutorial)
		}
	}

	if len(t.Firstname) > 0 {
		foundUser.PersonalInfo.Firstname = t.Firstname
	}

	if len(t.Lastname) > 0 {
		foundUser.PersonalInfo.Lastname = t.Lastname
	}

	if len(t.CompanyRole) > 0 {
		foundUser.PersonalInfo.Role = t.CompanyRole
	}

	if project.Environment == "cloud" {
		if len(t.EthInfo.Account) > 0 {
			log.Printf("[DEBUG] Should set ethinfo to %s", t.EthInfo)
			foundUser.EthInfo = t.EthInfo
		}

		username := foundUser.PublicProfile.GithubUsername
		creator, err := HandleAlgoliaCreatorSearch(ctx, username)

		// FIXME: If any of the parts below are updated, also update
		// the same field within Algolia itself.
		if err == nil {
			// Related to creators
			if len(t.CreatorDescription) > 0 {
				foundUser.PublicProfile.GithubBio = t.CreatorDescription
			}

			if len(t.CreatorUrl) > 0 {
				foundUser.PublicProfile.GithubUrl = t.CreatorUrl
			}

			if len(t.CreatorLocation) > 0 {
				foundUser.PublicProfile.GithubLocation = t.CreatorLocation
			}

			if len(t.CreatorSkills) > 0 {
				foundUser.PublicProfile.Skills = strings.Split(t.CreatorSkills, ",")
			}

			if len(t.CreatorWorkStatus) > 0 {
				foundUser.PublicProfile.WorkStatus = t.CreatorWorkStatus
			}

			if len(t.CreatorSocial) > 0 {
				foundUser.PublicProfile.Social = strings.Split(t.CreatorSocial, ",")
			}

			if len(t.SpecializedApps) > 0 {
				// FIXME: Update the user in algolia here. Currently just updating existing user
				for _, app := range t.SpecializedApps {
					found := false
					for _, currentApp := range creator.SpecializedApps {
						if currentApp.Name == app.Name {
							found = true
							break
						}
					}

					if !found {
						creator.SpecializedApps = append(creator.SpecializedApps, app)
					}
				}

				for _, creatorApp := range creator.SpecializedApps {
					// If not found in foundUser.PublicProfile
					found := false
					for _, userApp := range foundUser.PublicProfile.SpecializedApps {
						if userApp.Name == creatorApp.Name {
							found = true
							break
						}
					}

					if !found {
						foundUser.PublicProfile.SpecializedApps = append(foundUser.PublicProfile.SpecializedApps, creatorApp)
					}
				}

				//foundUser.PublicProfile.SpecializedApps = creator.SpecializedApps
			}
		}
	}

	if len(t.Suborgs) > 0 && foundUser.Id != userInfo.Id {
		//log.Printf("[DEBUG] Got suborg change: %s", t.Suborgs)
		// 1. Check if current users' active org is admin in same parent org as user
		// 2. Make sure the user should have access to suborg
		// 3. Make sure it's ONLY changing orgs based on parent org

		// Check which ones the current user has access to
		parentOrgId := userInfo.ActiveOrg.Id
		newSuborgs := []string{}
		for _, suborg := range t.Suborgs {
			if suborg == "REMOVE" {
				newSuborgs = append(newSuborgs, suborg)
				continue
			}

			found := false
			org, err := GetOrg(ctx, suborg)
			if err != nil {
				continue
			}

			if org.CreatorOrg != parentOrgId {
				continue
			}

			for _, userOrg := range org.Users {
				if userOrg.Id == userInfo.Id {
					found = true
					break
				}
			}

			if found {
				newSuborgs = append(newSuborgs, suborg)
			}
		}

		t.Suborgs = newSuborgs
		//log.Printf("[DEBUG] Valid suborgs: %s", t.Suborgs)

		// 244199a7-0009-44af-aefb-55da92334583
		// 583816e5-40ab-4212-8c7a-e54c8edd6b74

		addedOrgs := []string{}
		for _, suborg := range t.Suborgs {
			if suborg == "REMOVE" {
				continue
			}

			if ArrayContains(foundUser.Orgs, suborg) {
				log.Printf("[DEBUG] Skipping %s as it already exists", suborg)
				continue
			}

			if !ArrayContains(userInfo.Orgs, suborg) {
				log.Printf("[ERROR] Skipping org %s as user %s (%s) can't edit this one. Should never happen unless direct API usage.", suborg, userInfo.Username, userInfo.Id)
				continue
			}

			foundOrg, err := GetOrg(ctx, suborg)
			if err != nil {
				log.Printf("[WARNING] Failed to get suborg in user edit for %s (%s): %s", foundUser.Username, foundUser.Id, err)
				continue
			}

			// Slower but easier :)
			parsedOrgs := []string{foundOrg.CreatorOrg}
			for _, item := range foundOrg.ManagerOrgs {
				parsedOrgs = append(parsedOrgs, item.Id)
			}

			if !ArrayContains(parsedOrgs, userInfo.ActiveOrg.Id) {
				log.Printf("[ERROR] The Org %s (%s) SHOULD NOT BE ADDED for %s (%s): %s. This may indicate a test of the API, as the frontend shouldn't allow it.", foundOrg.Name, suborg, foundUser.Username, foundUser.Id, err)
				continue
			}

			addedOrgs = append(addedOrgs, suborg)
		}

		// After done, check if ANY of the users' orgs are suborgs of active parent org. If they are, remove.
		// Update: This piece runs anyway, in case the job is to REMOVE any suborg
		//if len(addedOrgs) > 0 {
		log.Printf("[DEBUG] Orgs to be added: %s. Existing: %s.", addedOrgs, foundUser.Orgs)

		// Removed for now due to multi-org chain deleting you from other org chains
		newUserOrgs := []string{}
		for _, suborg := range foundUser.Orgs {
			if suborg == userInfo.ActiveOrg.Id {
				newUserOrgs = append(newUserOrgs, suborg)
				continue
			}

			foundOrg, err := GetOrg(ctx, suborg)
			if err != nil {
				log.Printf("[WARNING] Failed to get suborg in user edit (2) for %s (%s): %s", foundUser.Username, foundUser.Id, err)
				newUserOrgs = append(newUserOrgs, suborg)
				continue
			}

			// Check if it has anything to do with the parent org, otherwise don't touch it
			// CreatorOrg, ManagerOrgs, Suborgs
			orgRelevancies := []string{foundOrg.CreatorOrg}
			for _, item := range foundOrg.ManagerOrgs {
				orgRelevancies = append(orgRelevancies, item.Id)
			}
			for _, item := range foundOrg.ChildOrgs {
				orgRelevancies = append(orgRelevancies, item.Id)
			}
			if !ArrayContains(orgRelevancies, userInfo.ActiveOrg.Id) {
				newUserOrgs = append(newUserOrgs, suborg)
				log.Printf("[DEBUG] Org %s (%s) is not relevant to parent org %s (%s). Skipping.", foundOrg.Name, foundOrg.Id, userInfo.ActiveOrg.Name, userInfo.ActiveOrg.Id)
				continue
			}

			// Slower but easier :)
			parsedOrgs := []string{foundOrg.CreatorOrg}
			for _, item := range foundOrg.ManagerOrgs {
				parsedOrgs = append(parsedOrgs, item.Id)
			}

			//if !ArrayContains(parsedOrgs, userInfo.ActiveOrg.Id) {
			if !ArrayContains(parsedOrgs, suborg) {
				if ArrayContains(t.Suborgs, suborg) {
					log.Printf("[DEBUG] Reappending org %s", suborg)
					newUserOrgs = append(newUserOrgs, suborg)
				} else {
					log.Printf("[DEBUG] Skipping org %s", suborg)
				}

				continue
			}

			log.Printf("[DEBUG] Should remove user %s (%s) from org %s if it doesn't exist in t.Suborgs", foundUser.Username, foundUser.Id, suborg)
			newUsers := []User{}
			for _, user := range foundOrg.Users {
				if user.Id == foundUser.Id {
					continue
				}

				newUsers = append(newUsers, user)
			}

			foundOrg.Users = newUsers
			err = SetOrg(ctx, *foundOrg, foundOrg.Id)
			if err != nil {
				log.Printf("[WARNING] Failed setting org when changing user access: %s", err)
			}

		}

		foundUser.Orgs = append(newUserOrgs, addedOrgs...)

		log.Printf("[DEBUG] New orgs for %s (%s) is %s", foundUser.Username, foundUser.Id, foundUser.Orgs)
	}

	err = SetUser(ctx, foundUser, orgUpdater)
	if err != nil {
		log.Printf("[WARNING] Error patching user %s: %s", foundUser.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false}`)))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true}`)))
}

func GenerateApikey(ctx context.Context, userInfo User) (User, error) {
	// Generate UUID
	// Set uuid to apikey in backend (update)
	userInfo.ApiKey = uuid.NewV4().String()
	err := SetApikey(ctx, userInfo)
	if err != nil {
		log.Printf("[WARNING] Failed updating apikey: %s", err)
		return userInfo, err
	}

	// Updating user
	log.Printf("[INFO] Adding apikey to user %s", userInfo.Username)
	err = SetUser(ctx, &userInfo, true)
	if err != nil {
		log.Printf("[WARNING] Failed updating users' apikey: %s", err)
		return userInfo, err
	}

	return userInfo, nil
}

func SetNewWorkflow(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in set new workflow: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to set new workflow: %s (%s)", user.Username, user.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error with body read: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var workflow Workflow
	err = json.Unmarshal(body, &workflow)
	if err != nil {
		log.Printf("Failed unmarshaling: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	workflow.ID = uuid.NewV4().String()
	workflow.Owner = user.Id
	workflow.Sharing = "private"
	user.ActiveOrg.Users = []UserMini{}
	workflow.ExecutingOrg = user.ActiveOrg
	workflow.OrgId = user.ActiveOrg.Id
	//log.Printf("TRIGGERS: %d", len(workflow.Triggers))

	ctx := GetContext(request)
	//err = increaseStatisticsField(ctx, "total_workflows", workflow.ID, 1, workflow.OrgId)
	//if err != nil {
	//	log.Printf("Failed to increase total workflows stats: %s", err)
	//}

	if len(workflow.Actions) == 0 {
		workflow.Actions = []Action{}
	}
	if len(workflow.Branches) == 0 {
		workflow.Branches = []Branch{}
	}
	if len(workflow.Triggers) == 0 {
		workflow.Triggers = []Trigger{}
	}
	if len(workflow.Errors) == 0 {
		workflow.Errors = []string{}
	}

	newActions := []Action{}
	for _, action := range workflow.Actions {
		if action.Environment == "" {
			//action.Environment = baseEnvironment

			// FIXME: Still necessary? This hinders hybrid mode cloud -> onprem
			//if project.Environment == "cloud" {
			//	action.Environment = "Cloud"
			//}

			action.IsValid = true
		}

		//action.LargeImage = ""
		newActions = append(newActions, action)
	}

	// Initialized without functions = adding a hello world node.
	if len(newActions) == 0 {
		//log.Printf("APPENDING NEW APP FOR NEW WORKFLOW")

		// Adds the Testing app if it's a new workflow
		workflowapps, err := GetPrioritizedApps(ctx, user)
		envName := "cloud"
		if project.Environment != "cloud" {
			workflowapps, err = GetAllWorkflowApps(ctx, 1000, 0)
			envName = "Shuffle"
		}

		//log.Printf("[DEBUG] Got %d apps. Err: %s", len(workflowapps), err)
		if err == nil {
			environments, err := GetEnvironments(ctx, user.ActiveOrg.Id)
			if err == nil {
				for _, env := range environments {
					if env.Default {
						envName = env.Name
						break
					}
				}
			}

			for _, item := range workflowapps {
				//log.Printf("NAME: %s", item.Name)
				if (item.Name == "Shuffle Tools" || item.Name == "Shuffle-Tools") && item.AppVersion == "1.2.0" {
					//nodeId := "40447f30-fa44-4a4f-a133-4ee710368737"
					nodeId := uuid.NewV4().String()
					workflow.Start = nodeId
					newAction := Action{
						Label:       "Change Me",
						Name:        "repeat_back_to_me",
						Environment: envName,
						Parameters: []WorkflowAppActionParameter{
							WorkflowAppActionParameter{
								Name:      "call",
								Value:     "Hello world",
								Example:   "Repeating: Hello World",
								Multiline: true,
							},
						},
						Priority:    0,
						Errors:      []string{},
						ID:          nodeId,
						IsValid:     true,
						IsStartNode: true,
						Sharing:     true,
						PrivateID:   "",
						SmallImage:  "",
						AppName:     item.Name,
						AppVersion:  item.AppVersion,
						AppID:       item.ID,
						LargeImage:  item.LargeImage,
					}
					newAction.Position = Position{
						X: 449.5,
						Y: 446,
					}

					newActions = append(newActions, newAction)

					break
				}
			}
		}
	} else {
		log.Printf("[INFO] Has %d actions already", len(newActions))
		// FIXME: Check if they require authentication and if they exist locally
		//log.Printf("\n\nSHOULD VALIDATE AUTHENTICATION")
		//AuthenticationId string `json:"authentication_id,omitempty" datastore:"authentication_id"`
		//allAuths, err := GetAllWorkflowAppAuth(ctx, user.ActiveOrg.Id)
		//if err == nil {
		//	log.Printf("AUTH: %s", allAuths)
		//	for _, action := range newActions {
		//		log.Printf("ACTION: %s", action)
		//	}
		//}
	}

	workflow.Actions = []Action{}
	for _, item := range workflow.Actions {
		oldId := item.ID
		sourceIndexes := []int{}
		destinationIndexes := []int{}
		for branchIndex, branch := range workflow.Branches {
			if branch.SourceID == oldId {
				sourceIndexes = append(sourceIndexes, branchIndex)
			}

			if branch.DestinationID == oldId {
				destinationIndexes = append(destinationIndexes, branchIndex)
			}
		}

		item.ID = uuid.NewV4().String()
		for _, index := range sourceIndexes {
			workflow.Branches[index].SourceID = item.ID
		}

		for _, index := range destinationIndexes {
			workflow.Branches[index].DestinationID = item.ID
		}

		newActions = append(newActions, item)
	}

	newTriggers := []Trigger{}
	for _, item := range workflow.Triggers {
		oldId := item.ID
		sourceIndexes := []int{}
		destinationIndexes := []int{}
		for branchIndex, branch := range workflow.Branches {
			if branch.SourceID == oldId {
				sourceIndexes = append(sourceIndexes, branchIndex)
			}

			if branch.DestinationID == oldId {
				destinationIndexes = append(destinationIndexes, branchIndex)
			}
		}

		item.ID = uuid.NewV4().String()
		for _, index := range sourceIndexes {
			workflow.Branches[index].SourceID = item.ID
		}

		for _, index := range destinationIndexes {
			workflow.Branches[index].DestinationID = item.ID
		}

		item.Status = "uninitialized"
		newTriggers = append(newTriggers, item)
	}

	newSchedules := []Schedule{}
	for _, item := range workflow.Schedules {
		item.Id = uuid.NewV4().String()
		newSchedules = append(newSchedules, item)
	}

	timeNow := int64(time.Now().Unix())
	workflow.Actions = newActions
	workflow.Triggers = newTriggers
	workflow.Schedules = newSchedules
	workflow.IsValid = true
	workflow.Configuration.ExitOnError = false
	workflow.Created = timeNow

	workflowjson, err := json.Marshal(workflow)
	if err != nil {
		log.Printf("Failed workflow json setting marshalling: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	err = SetWorkflow(ctx, workflow, workflow.ID)
	if err != nil {
		log.Printf("[WARNING] Failed setting workflow: %s (Set workflow)", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Cleans up cache for the users
	org, err := GetOrg(ctx, user.ActiveOrg.Id)
	if err == nil {
		//log.Printf("Getting Org workflows")

		workflows, err := GetAllWorkflowsByQuery(ctx, user)
		if err == nil {
			updated := false
			for tutorialIndex, tutorial := range org.Tutorials {
				if tutorial.Name == "Discover Usecases" {
					org.Tutorials[tutorialIndex].Description = fmt.Sprintf("%d workflows created. Find more using Workflow Templates or public Workflows.", len(workflows)+1)
					if len(workflows) > 0 {
						org.Tutorials[tutorialIndex].Done = true
						//org.Tutorials[tutorialIndex].Link = "/search?tab=workflows"
						org.Tutorials[tutorialIndex].Link = "/usecases"
					}

					updated = true
					break
				}
			}

			if updated {
				SetOrg(ctx, *org, org.Id)
			}
		} else {
			log.Printf("[ERROR] Failed getting workflows during new workflow creation for updating stats: %s", err)
		}

		for _, loopUser := range org.Users {
			cacheKey := fmt.Sprintf("%s_workflows", loopUser.Id)
			DeleteCache(ctx, cacheKey)
		}
	} else {
		cacheKey := fmt.Sprintf("%s_workflows", user.Id)
		DeleteCache(ctx, cacheKey)
	}

	log.Printf("[INFO] Saved new workflow %s with name %s", workflow.ID, workflow.Name)

	resp.WriteHeader(200)
	resp.Write(workflowjson)
}

// Saves a workflow to an ID
func SaveWorkflow(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, userErr := HandleApiAuthentication(resp, request)
	if userErr != nil {
		log.Printf("[WARNING] Api authentication failed in save workflow: %s", userErr)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to save workflow (2): %s (%s)", user.Username, user.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")

	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
		if strings.Contains(fileId, "?") {
			fileId = strings.Split(fileId, "?")[0]
		}
	}

	if len(fileId) != 36 {
		log.Printf(`[WARNING] Workflow ID %s is not valid`, fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow ID to save is not valid"}`))
		return
	}

	// Here to check access rights
	ctx := GetContext(request)
	tmpworkflow, err := GetWorkflow(ctx, fileId)
	if err != nil {
		log.Printf("[WARNING] Failed getting the workflow %s locally (save workflow): %s", fileId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	workflow := Workflow{}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Failed workflow body read: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	err = json.Unmarshal([]byte(body), &workflow)
	if err != nil {
		//log.Printf(string(body))
		log.Printf("[ERROR] Failed workflow unmarshaling (save): %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
		return
	}

	type PublicCheck struct {
		UserEditing bool   `json:"user_editing"`
		Public      bool   `json:"public"`
		Owner       string `json:"owner"`
	}

	correctUser := false
	if user.Id != tmpworkflow.Owner || tmpworkflow.Public == true {
		if tmpworkflow.Public {
			// FIXME:
			// If the user Id is part of the creator: DONT update this way.
			// /users/creators/username
			// Just making sure
			if project.Environment == "cloud" {
				//algoliaUser, err := HandleAlgoliaCreatorSearch(ctx, username)
				algoliaUser, err := HandleAlgoliaCreatorSearch(ctx, tmpworkflow.ID)
				if err != nil {
					log.Printf("[WARNING] User with ID %s for Workflow %s could not be found (workflow update): %s", user.Id, tmpworkflow.ID, err)

					// Check if current user is one of the few allowed
					// This can only happen if the workflow doesn't already have an owner
					//log.Printf("CUR USER: %s\n\n%s", user.PublicProfile, os.Getenv("GITHUB_USER_ALLOWLIST"))
					allowList := os.Getenv("GITHUB_USER_ALLOWLIST")
					found := false
					if user.PublicProfile.Public && len(allowList) > 0 {
						allowListSplit := strings.Split(allowList, ",")
						for _, username := range allowListSplit {
							if username == user.PublicProfile.GithubUsername {
								algoliaUser, err = HandleAlgoliaCreatorSearch(ctx, user.PublicProfile.GithubUsername)
								if err != nil {
									log.Printf("New error: %s", err)
								}

								found = true
								break
							}

						}

					}

					if !found {
						resp.WriteHeader(401)
						resp.Write([]byte(`{"success": false}`))
						return
					}
				}

				wf2 := PublicCheck{}
				err = json.Unmarshal([]byte(body), &wf2)
				if err != nil {
					log.Printf("[ERROR] Failed workflow unmarshaling (save - 2): %s", err)
				}

				if algoliaUser.ObjectID == user.Id || ArrayContains(algoliaUser.Synonyms, user.Id) {
					log.Printf("[WARNING] User %s (%s) has access to edit %s! Keep it public!!", user.Username, user.Id, workflow.ID)

					// Means the owner is using the workflow for their org
					if wf2.UserEditing == false {
						correctUser = false
					} else {
						correctUser = true
						tmpworkflow.Public = true
						workflow.Public = true
					}
				}
			}

			// FIX: Should check if this workflow has already been saved?
			if !correctUser {
				log.Printf("[INFO] User %s is saving the public workflow %s", user.Username, tmpworkflow.ID)
				workflow = *tmpworkflow
				workflow.PublishedId = workflow.ID
				workflow.ID = uuid.NewV4().String()
				workflow.Public = false
				workflow.Owner = user.Id
				workflow.Org = []OrgMini{
					user.ActiveOrg,
				}
				workflow.ExecutingOrg = user.ActiveOrg
				workflow.OrgId = user.ActiveOrg.Id
				workflow.PreviouslySaved = false

				newTriggers := []Trigger{}
				changedIds := map[string]string{}
				for _, trigger := range workflow.Triggers {
					log.Printf("TriggerID: %s", trigger.ID)
					newId := uuid.NewV4().String()
					trigger.Environment = "cloud"

					hookAuth := ""
					customResponse := ""
					for paramIndex, param := range trigger.Parameters {
						if param.Name == "url" {
							trigger.Parameters[paramIndex].Value = fmt.Sprintf("https://shuffler.io/api/v1/hooks/webhook_%s", newId)
						}

						if param.Name == "auth_headers" {
							hookAuth = param.Value
						}

						if param.Name == "custom_response_body" {
							customResponse = param.Value
						}
					}

					if trigger.TriggerType != "SCHEDULE" {

						trigger.Status = "running"

						if trigger.TriggerType == "WEBHOOK" {
							hook := Hook{
								Id:        newId,
								Start:     workflow.Start,
								Workflows: []string{workflow.ID},
								Info: Info{
									Name:        trigger.Name,
									Description: trigger.Description,
									Url:         fmt.Sprintf("https://shuffler.io/api/v1/hooks/webhook_%s", newId),
								},
								Type:   "webhook",
								Owner:  user.Username,
								Status: "running",
								Actions: []HookAction{
									HookAction{
										Type:  "workflow",
										Name:  trigger.Name,
										Id:    workflow.ID,
										Field: "",
									},
								},
								Running:        true,
								OrgId:          user.ActiveOrg.Id,
								Environment:    "cloud",
								Auth:           hookAuth,
								CustomResponse: customResponse,
							}

							log.Printf("[DEBUG] Starting hook %s for user %s (%s) during Workflow Save for %s", hook.Id, user.Username, user.Id, workflow.ID)
							err = SetHook(ctx, hook)
							if err != nil {
								log.Printf("[WARNING] Failed setting hook during workflow copy of %s: %s", workflow.ID, err)
								resp.WriteHeader(401)
								resp.Write([]byte(`{"success": false}`))
								return
							}
						}
					}

					changedIds[trigger.ID] = newId

					trigger.ID = newId
					//log.Printf("New id for %s: %s", trigger.TriggerType, trigger.ID)
					newTriggers = append(newTriggers, trigger)
				}

				newBranches := []Branch{}
				for _, branch := range workflow.Branches {
					for key, value := range changedIds {
						if branch.SourceID == key {
							branch.SourceID = value
						}

						if branch.DestinationID == key {
							branch.DestinationID = value
						}
					}

					newBranches = append(newBranches, branch)
				}

				workflow.Branches = newBranches
				workflow.Triggers = newTriggers

				err = SetWorkflow(ctx, workflow, workflow.ID)
				if err != nil {
					log.Printf("[WARNING] Failed saving NEW version of public %s for user %s: %s", tmpworkflow.ID, user.Username, err)
					resp.WriteHeader(401)
					resp.Write([]byte(`{"success": false}`))
					return
				}
				org, err := GetOrg(ctx, user.ActiveOrg.Id)
				if err != nil {
					log.Printf("[WARNING] Failed getting org for cache release for public wf: %s", err)
				} else {
					for _, loopUser := range org.Users {
						DeleteCache(ctx, fmt.Sprintf("%s_workflows", loopUser.Id))
						DeleteCache(ctx, fmt.Sprintf("apps_%s", loopUser.Id))
						DeleteCache(ctx, fmt.Sprintf("user_%s", loopUser.Id))
					}

					// Activate all that aren't already there
					changed := false
					for _, action := range workflow.Actions {
						//log.Printf("App: %s, Public: %s", action.AppID, action.Public)
						if !ArrayContains(org.ActiveApps, action.AppID) {
							org.ActiveApps = append(org.ActiveApps, action.AppID)
							changed = true
						}
					}

					if changed {
						err = SetOrg(ctx, *org, org.Id)
						if err != nil {
							log.Printf("[ERROR] Failed updating active app list for org %s (%s): %s", org.Name, org.Id, err)
						} else {
							DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
							DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-100"))
							DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-500"))
							DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-1000"))
							DeleteCache(ctx, "all_apps")
							DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
							DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))
						}
					}
				}

				resp.WriteHeader(200)
				resp.Write([]byte(fmt.Sprintf(`{"success": true, "new_id": "%s"}`, workflow.ID)))
				return
			}
		} else if tmpworkflow.OrgId == user.ActiveOrg.Id && user.Role == "admin" {
			log.Printf("[AUDIT] User %s is accessing workflow %s as admin (save workflow)", user.Username, tmpworkflow.ID)
			workflow.ID = tmpworkflow.ID
		} else {
			log.Printf("[AUDIT] Wrong user (%s) for workflow %s (save)", user.Username, tmpworkflow.ID)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	} else {

		if workflow.Public {
			log.Printf("[WARNING] Rolling back public as the user set it to true themselves")
			workflow.Public = false
		}

		if len(workflow.PublishedId) > 0 {
			log.Printf("[INFO] Workflow %s has the published ID %s", workflow.ID, workflow.PublishedId)
		}
	}

	if fileId != workflow.ID {
		log.Printf("[WARNING] Path and request ID are not matching in workflow save: %s != %s.", fileId, workflow.ID)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if len(workflow.Name) == 0 {
		log.Printf("[WARNING] Can't save workflow without a name.")
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow needs a name"}`))
		return
	}

	if len(workflow.Actions) == 0 {
		log.Printf("[WARNING] Can't save workflow without a single action.")
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow needs at least one action"}`))
		return
	}

	// Resetting subflows as they shouldn't be entirely saved. Used just for imports/exports
	if len(workflow.Subflows) > 0 {
		log.Printf("[DEBUG] Got %d subflows saved in %s (to be saved and removed)", len(workflow.Subflows), workflow.ID)

		for _, subflow := range workflow.Subflows {
			SetWorkflow(ctx, subflow, subflow.ID)
		}

		workflow.Subflows = []Workflow{}
	}

	if workflow.Status != "test" && workflow.Status != "production" {
		workflow.Status = "test"
		log.Printf("[DEBUG] Defaulted workflow status to %s. Alternative: prod", workflow.Status)
	}

	if strings.ToLower(workflow.Status) == "prod" {
		workflow.Status = "production"
	}

	workflow.Subflows = []Workflow{}
	if len(workflow.DefaultReturnValue) > 0 && len(workflow.DefaultReturnValue) < 200 {
		log.Printf("[INFO] Set default return value to on failure to (%s): %s", workflow.ID, workflow.DefaultReturnValue)
		//workflow.DefaultReturnValue
	}

	log.Printf("[INFO] Saving workflow '%s' with %d action(s) and %d trigger(s)", workflow.Name, len(workflow.Actions), len(workflow.Triggers))

	if len(user.ActiveOrg.Id) > 0 {
		if len(workflow.ExecutingOrg.Id) == 0 {
			log.Printf("[INFO] Setting executing org for workflow to %s", user.ActiveOrg.Id)
			user.ActiveOrg.Users = []UserMini{}
			workflow.ExecutingOrg = user.ActiveOrg
		}

		//if len(workflow.Org) == 0 {
		//	user.ActiveOrg.Users = []UserMini{}
		//	//workflow.Org = user.ActiveOrg
		//}

		if len(workflow.OrgId) == 0 {
			workflow.OrgId = user.ActiveOrg.Id
		}
	}

	newActions := []Action{}
	allNodes := []string{}
	workflow.Categories = Categories{}

	environments := []Environment{
		Environment{
			Name:       "Cloud",
			Type:       "cloud",
			Archived:   false,
			Registered: true,
			Default:    false,
			OrgId:      user.ActiveOrg.Id,
			Id:         uuid.NewV4().String(),
		},
	}

	//if project.Environment != "cloud" {
	environments, err = GetEnvironments(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Failed getting environments for org %s", user.ActiveOrg.Id)
		environments = []Environment{}
	}
	//}

	//log.Printf("ENVIRONMENTS: %s", environments)
	defaultEnv := ""
	for _, env := range environments {
		if env.Default {
			defaultEnv = env.Name
			break
		}
	}

	if defaultEnv == "" {
		if project.Environment == "cloud" {
			defaultEnv = "Cloud"
		} else {
			defaultEnv = "Shuffle"
		}
	}

	orgUpdated := false
	startnodeFound := false
	workflowapps, apperr := GetPrioritizedApps(ctx, user)
	newOrgApps := []string{}
	org := &Org{}
	for _, action := range workflow.Actions {
		if action.SourceWorkflow != workflow.ID && len(action.SourceWorkflow) > 0 {
			continue
		}

		allNodes = append(allNodes, action.ID)
		if workflow.Start == action.ID {
			//log.Printf("[INFO] FOUND STARTNODE %d", workflow.Start)
			startnodeFound = true
			action.IsStartNode = true
		}

		if len(action.Errors) > 0 || !action.IsValid {
			action.IsValid = true
			action.Errors = []string{}
		}

		if action.ExecutionDelay > 86400 {
			parsedError := fmt.Sprintf("Max execution delay for an action is 86400 (1 day)")
			if !ArrayContains(workflow.Errors, parsedError) {
				workflow.Errors = append(workflow.Errors, parsedError)
			}

			action.ExecutionDelay = 86400
		}

		if action.Environment == "" {
			if project.Environment == "cloud" {
				action.Environment = defaultEnv
			} else {
				if len(environments) > 0 {
					for _, env := range environments {
						if !env.Archived && env.Default {
							//log.Printf("FOUND ENV %s", env)
							action.Environment = env.Name
							break
						}
					}
				}

				if action.Environment == "" {
					action.Environment = defaultEnv
				}

				action.IsValid = true
			}
		} else {
			warned := []string{}
			found := false
			for _, env := range environments {
				if env.Name == action.Environment {
					found = true
					if env.Archived {
						log.Printf("[DEBUG] Environment %s is archived. Changing to default.")
						action.Environment = defaultEnv
					}

					break
				}
			}

			if !found {
				if ArrayContains(warned, action.Environment) {
					log.Printf("[DEBUG] Environment %s isn't available. Changing to default.", action.Environment)
					warned = append(warned, action.Environment)
				}

				action.Environment = defaultEnv
			}
		}

		// Fixing apps with bad IDs. This can happen a lot because of
		// autogeneration of app IDs, and export/imports of workflows
		idFound := false
		nameVersionFound := false
		nameFound := false
		discoveredApp := WorkflowApp{}
		for _, innerApp := range workflowapps {
			if innerApp.ID == action.AppID {
				discoveredApp = innerApp
				//log.Printf("[INFO] ID, Name AND version for %s:%s (%s) was FOUND", action.AppName, action.AppVersion, action.AppID)
				action.Sharing = innerApp.Sharing
				action.Public = innerApp.Public
				action.Generated = innerApp.Generated
				action.ReferenceUrl = innerApp.ReferenceUrl
				idFound = true
				break
			}
		}

		if !idFound {
			for _, innerApp := range workflowapps {
				if innerApp.Name == action.AppName && innerApp.AppVersion == action.AppVersion {
					discoveredApp = innerApp

					action.AppID = innerApp.ID
					action.Sharing = innerApp.Sharing
					action.Public = innerApp.Public
					action.Generated = innerApp.Generated
					action.ReferenceUrl = innerApp.ReferenceUrl
					nameVersionFound = true
					break
				}
			}
		}

		if !idFound {
			for _, innerApp := range workflowapps {
				if innerApp.Name == action.AppName {
					discoveredApp = innerApp

					action.AppID = innerApp.ID
					action.Sharing = innerApp.Sharing
					action.Public = innerApp.Public
					action.Generated = innerApp.Generated
					action.ReferenceUrl = innerApp.ReferenceUrl

					nameFound = true
					break
				}
			}
		}

		if !idFound {
			if nameVersionFound {
			} else if nameFound {
			} else {
				log.Printf("[WARNING] ID, Name AND version for %s:%s (%s) was NOT found", action.AppName, action.AppVersion, action.AppID)
				handled := false

				if project.Environment == "cloud" {
					appid, err := HandleAlgoliaAppSearch(ctx, action.AppName)
					if err == nil && len(appid.ObjectID) > 0 {
						//log.Printf("[INFO] Found NEW appid %s for app %s", appid, action.AppName)
						tmpApp, err := GetApp(ctx, appid.ObjectID, user, false)
						if err == nil {
							handled = true
							action.AppID = tmpApp.ID
							newOrgApps = append(newOrgApps, action.AppID)

							workflowapps = append(workflowapps, *tmpApp)
						}
					} else {
						log.Printf("[WARNING] Failed finding name %s in Algolia", action.AppName)
					}
				}

				if !handled {
					action.IsValid = false
					action.Errors = []string{fmt.Sprintf("Couldn't find app %s:%s", action.AppName, action.AppVersion)}
				}
			}
		}

		if !action.IsValid && len(action.Errors) > 0 {
			log.Printf("[INFO] Node %s is invalid and needs to be remade. Errors: %s", action.Label, strings.Join(action.Errors, "\n"))

			//if workflow.PreviouslySaved {
			//	resp.WriteHeader(401)
			//	resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Node %s is invalid and needs to be remade."}`, action.Label)))
			//	return
			//}
			//action.IsValid = true
			//action.Errors = []string{}
		}

		workflow.Categories = HandleCategoryIncrease(workflow.Categories, action, workflowapps)
		newActions = append(newActions, action)

		// FIXMe: Should be authenticated first?
		if len(discoveredApp.Categories) > 0 {
			category := discoveredApp.Categories[0]

			if org.Id == "" {
				org, err = GetOrg(ctx, user.ActiveOrg.Id)
				if err != nil {
					log.Printf("[WARNING] Failed getting org: %s", err)
					continue
				}
			}

			if strings.ToLower(category) == "siem" && org.SecurityFramework.SIEM.ID == "" {
				org.SecurityFramework.SIEM.Name = discoveredApp.Name
				org.SecurityFramework.SIEM.Description = discoveredApp.Description
				org.SecurityFramework.SIEM.ID = discoveredApp.ID
				org.SecurityFramework.SIEM.LargeImage = discoveredApp.LargeImage

				orgUpdated = true
			} else if strings.ToLower(category) == "network" && org.SecurityFramework.Network.ID == "" {
				org.SecurityFramework.Network.Name = discoveredApp.Name
				org.SecurityFramework.Network.Description = discoveredApp.Description
				org.SecurityFramework.Network.ID = discoveredApp.ID
				org.SecurityFramework.Network.LargeImage = discoveredApp.LargeImage

				orgUpdated = true
			} else if strings.ToLower(category) == "edr" || strings.ToLower(category) == "edr & av" && org.SecurityFramework.EDR.ID == "" {
				org.SecurityFramework.EDR.Name = discoveredApp.Name
				org.SecurityFramework.EDR.Description = discoveredApp.Description
				org.SecurityFramework.EDR.ID = discoveredApp.ID
				org.SecurityFramework.EDR.LargeImage = discoveredApp.LargeImage

				orgUpdated = true
			} else if strings.ToLower(category) == "cases" && org.SecurityFramework.Cases.ID == "" {
				org.SecurityFramework.Cases.Name = discoveredApp.Name
				org.SecurityFramework.Cases.Description = discoveredApp.Description
				org.SecurityFramework.Cases.ID = discoveredApp.ID
				org.SecurityFramework.Cases.LargeImage = discoveredApp.LargeImage

				orgUpdated = true
			} else if strings.ToLower(category) == "iam" && org.SecurityFramework.IAM.ID == "" {
				org.SecurityFramework.IAM.Name = discoveredApp.Name
				org.SecurityFramework.IAM.Description = discoveredApp.Description
				org.SecurityFramework.IAM.ID = discoveredApp.ID
				org.SecurityFramework.IAM.LargeImage = discoveredApp.LargeImage

				orgUpdated = true
			} else if strings.ToLower(category) == "assets" && org.SecurityFramework.Assets.ID == "" {
				log.Printf("Setting assets?")
				org.SecurityFramework.Assets.Name = discoveredApp.Name
				org.SecurityFramework.Assets.Description = discoveredApp.Description
				org.SecurityFramework.Assets.ID = discoveredApp.ID
				org.SecurityFramework.Assets.LargeImage = discoveredApp.LargeImage

				orgUpdated = true
			} else if strings.ToLower(category) == "intel" && org.SecurityFramework.Intel.ID == "" {
				org.SecurityFramework.Intel.Name = discoveredApp.Name
				org.SecurityFramework.Intel.Description = discoveredApp.Description
				org.SecurityFramework.Intel.ID = discoveredApp.ID
				org.SecurityFramework.Intel.LargeImage = discoveredApp.LargeImage

				orgUpdated = true
			} else if strings.ToLower(category) == "comms" && org.SecurityFramework.Communication.ID == "" {
				org.SecurityFramework.Communication.Name = discoveredApp.Name
				org.SecurityFramework.Communication.Description = discoveredApp.Description
				org.SecurityFramework.Communication.ID = discoveredApp.ID
				org.SecurityFramework.Communication.LargeImage = discoveredApp.LargeImage

				orgUpdated = true
			} else {
				//log.Printf("[WARNING] No handler for type %s in app framework", category)
			}

		}
	}

	if !startnodeFound {
		log.Printf("[WARNING] No startnode found during save of %s!!", workflow.ID)
	}

	// Automatically adding new apps
	if len(newOrgApps) > 0 {
		log.Printf("[WARNING] Adding new apps to org: %s", newOrgApps)

		if org.Id == "" {
			org, err = GetOrg(ctx, user.ActiveOrg.Id)
			if err != nil {
				log.Printf("[WARNING] Failed getting org during new app update for %s: %s", user.ActiveOrg.Id, err)
			}
		}

		if org.Id != "" {
			added := false
			for _, newApp := range newOrgApps {
				if !ArrayContains(org.ActiveApps, newApp) {
					org.ActiveApps = append(org.ActiveApps, newApp)
					added = true
				}
			}

			if added {
				orgUpdated = true
				//err = SetOrg(ctx, *org, org.Id)
				//if err != nil {
				//	log.Printf("[WARNING] Failed setting org when autoadding apps on save: %s", err)
				//} else {
				DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
				DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-100"))
				DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-500"))
				DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-1000"))
				DeleteCache(ctx, "all_apps")
				DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
				DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))
			}
			//}
		}
	}

	workflow.Actions = newActions

	newTriggers := []Trigger{}
	for _, trigger := range workflow.Triggers {
		if trigger.SourceWorkflow != workflow.ID && len(trigger.SourceWorkflow) > 0 {
			continue
		}

		// Check if it's actually running
		if trigger.TriggerType == "SCHEDULE" && trigger.Status != "uninitialized" {
			schedule, err := GetSchedule(ctx, trigger.ID)
			if err != nil {
				trigger.Status = "stopped"
			} else if schedule.Id == "" {
				trigger.Status = "stopped"
			}
		} else if trigger.TriggerType == "SUBFLOW" {
			for _, param := range trigger.Parameters {
				if param.Name == "workflow" {
					// Validate workflow exists
					_, err := GetWorkflow(ctx, param.Value)
					if err != nil {
						parsedError := fmt.Sprintf("Workflow %s in Subflow %s (%s) doesn't exist", workflow.ID, trigger.Label, trigger.ID)
						if !ArrayContains(workflow.Errors, parsedError) {
							workflow.Errors = append(workflow.Errors, parsedError)
						}

						log.Printf("[ERROR] Couldn't find subflow %s for workflow %s (%s). NOT setting to self as failover for now, and trusting authentication system instead.", param.Value, workflow.Name, workflow.ID)
						//trigger.Parameters[paramIndex].Value = workflow.ID
					}
				}

				//if len(param.Value) == 0 && param.Name != "argument" {
				// FIXME: No longer necessary to use the org's users' actual APIkey
				// Instead, this is replaced during runtime to use the executions' key
				/*
					if param.Name == "user_apikey" {
						apikey := ""
						if len(user.ApiKey) > 0 {
							apikey = user.ApiKey
						} else {
							user, err = GenerateApikey(ctx, user)
							if err != nil {
								workflow.IsValid = false
								workflow.Errors = []string{"Trigger is missing a parameter: %s", param.Name}

								log.Printf("[DEBUG] No type specified for subflow node")

								if workflow.PreviouslySaved {
									resp.WriteHeader(401)
									resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Trigger %s is missing the parameter %s"}`, trigger.Label, param.Name)))
									return
								}
							}

							apikey = user.ApiKey
						}

						log.Printf("[INFO] Set apikey in subflow trigger for user during save")
						if len(apikey) > 0 {
							trigger.Parameters[index].Value = apikey
						}
					}
				*/
				//} else {

				//	workflow.IsValid = false
				//	workflow.Errors = []string{"Trigger is missing a parameter: %s", param.Name}

				//	log.Printf("[WARNING] No type specified for user input node")
				//	if workflow.PreviouslySaved {
				//		resp.WriteHeader(401)
				//		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Trigger %s is missing the parameter %s"}`, trigger.Label, param.Name)))
				//		return
				//	}
				//}
				//}
			}
		} else if trigger.TriggerType == "WEBHOOK" {
			if trigger.Status != "uninitialized" && trigger.Status != "stopped" {
				hook, err := GetHook(ctx, trigger.ID)
				if err != nil {
					log.Printf("[WARNING] Failed getting webhook %s (%s)", trigger.ID, trigger.Status)
					trigger.Status = "stopped"
				} else if hook.Id == "" {
					trigger.Status = "stopped"
				}
			}

			//log.Printf("WEBHOOK: %d", len(trigger.Parameters))
			if len(trigger.Parameters) < 2 {
				log.Printf("[WARNING] Issue with parameters in webhook %s - missing params", trigger.ID)
			} else {
				if !strings.Contains(trigger.Parameters[0].Value, trigger.ID) {
					log.Printf("[INFO] Fixing webhook URL for %s", trigger.ID)
					baseUrl := "https://shuffler.io"
					if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
						baseUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
					}

					if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
						baseUrl = os.Getenv("SHUFFLE_CLOUDRUN_URL")
					}

					if project.Environment != "cloud" {
						baseUrl = "http://localhost:3001"
					}

					newTriggerName := fmt.Sprintf("webhook_%s", trigger.ID)
					trigger.Parameters[0].Value = fmt.Sprintf("%s/api/v1/hooks/%s", baseUrl, newTriggerName)
					trigger.Parameters[1].Value = newTriggerName
				}
			}
		} else if trigger.TriggerType == "USERINPUT" {
			// E.g. check email
			sms := ""
			email := ""
			subflow := ""
			triggerType := ""
			triggerInformation := ""
			for _, item := range trigger.Parameters {
				if item.Name == "alertinfo" {
					triggerInformation = item.Value
				} else if item.Name == "type" {
					triggerType = item.Value
				} else if item.Name == "email" {
					email = item.Value
				} else if item.Name == "sms" {
					sms = item.Value
				} else if item.Name == "subflow" {
					subflow = item.Value
				}
			}

			_ = subflow

			if len(triggerType) == 0 {
				log.Printf("[DEBUG] No type specified for user input node")
				if workflow.PreviouslySaved {
					//resp.WriteHeader(401)
					//resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "No contact option specified in user input"}`)))
					//return
				}
			}

			// FIXME: This is not the right time to send them, BUT it's well served for testing. Save -> send email / sms
			_ = triggerInformation
			if strings.Contains(triggerType, "email") {
				if email == "test@test.com" {
					log.Printf("Email isn't specified during save.")
					if workflow.PreviouslySaved {
						resp.WriteHeader(401)
						resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Email field in user input can't be empty"}`)))
						return
					}
				}

				//log.Printf("[DEBUG] Should send email to %s during execution.", email)
			}

			if strings.Contains(triggerType, "sms") {
				if sms == "0000000" {
					log.Printf("Email isn't specified during save.")
					if workflow.PreviouslySaved {
						resp.WriteHeader(401)
						resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "SMS field in user input can't be empty"}`)))
						return
					}
				}

				log.Printf("[DEBUG] Should send SMS to %s during execution.", sms)
			}

			// Removed all checks as this is handled in the shuffle-subflow app
			if strings.Contains(triggerType, "subflow") {
				//if strings.Contains(subflow, "/") {
				//	subflowSplit := strings.Split(subflow, "/")
				//	subflow = subflowSplit[len(subflowSplit)-1]
				//}

				//if len(subflow) != 36 {
				//	log.Printf("[WARNING] Subflow isn't specified!")
				//	if workflow.PreviouslySaved {
				//		resp.WriteHeader(401)
				//		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Subflow in User Input Trigger isn't specified"}`)))
				//		return
				//	}
				//}

				//log.Printf("[DEBUG] Should run subflow with workflow %s during execution.", subflow)
			}
		}

		allNodes = append(allNodes, trigger.ID)
		newTriggers = append(newTriggers, trigger)
	}

	newComments := []Comment{}
	for _, comment := range workflow.Comments {
		if comment.Height < 50 {
			comment.Height = 150
		}

		if comment.Width < 50 {
			comment.Height = 150
		}

		if len(comment.BackgroundColor) == 0 {
			comment.BackgroundColor = "#1f2023"
		}

		if len(comment.Color) == 0 {
			comment.Color = "#ffffff"
		}

		comment.Position.X = float64(comment.Position.X)
		comment.Position.Y = float64(comment.Position.Y)

		newComments = append(newComments, comment)
	}

	workflow.Comments = newComments
	workflow.Triggers = newTriggers

	if len(workflow.Actions) == 0 {
		workflow.Actions = []Action{}
	}
	if len(workflow.Branches) == 0 {
		workflow.Branches = []Branch{}
	}
	if len(workflow.Triggers) == 0 {
		workflow.Triggers = []Trigger{}
	}
	if len(workflow.Errors) == 0 {
		workflow.Errors = []string{}
	}
	if len(workflow.Comments) == 0 {
		workflow.Comments = []Comment{}
	}

	//log.Printf("PRE VARIABLES")
	for _, variable := range workflow.WorkflowVariables {
		if len(variable.Value) == 0 {
			log.Printf("[WARNING] Variable %s is empty!", variable.Name)
			workflow.Errors = append(workflow.Errors, fmt.Sprintf("Variable %s is empty!", variable.Name))
			//resp.WriteHeader(401)
			//resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Variable %s can't be empty"}`, variable.Name)))
			//return
			//} else {
			//	log.Printf("VALUE OF VAR IS %s", variable.Value)
		}
	}

	if len(workflow.ExecutionVariables) > 0 {
		log.Printf("[INFO] Found %d runtime variable(s) for workflow %s", len(workflow.ExecutionVariables), workflow.ID)
	}

	if len(workflow.WorkflowVariables) > 0 {
		log.Printf("[INFO] Found %d workflow variable(s) for workflow %s", len(workflow.WorkflowVariables), workflow.ID)
	}

	// Nodechecks
	foundNodes := []string{}
	for _, node := range allNodes {
		for _, branch := range workflow.Branches {
			if node == branch.DestinationID || node == branch.SourceID {
				foundNodes = append(foundNodes, node)
				break
			}

			// Check if source parent
		}
	}

	// Making sure to check if infinite loops exist
	// This is now done during execution. Should be done on frontend & backend too
	/*
		for _, action := range workflow.Actions {
			//log.Printf("[INFO] Checking childnodes for action %s (%s)", action.Label, action.ID)
			childNodes := FindChildNodes(WorkflowExecution{Workflow: workflow}, action.ID, []string{}, []string{})
			log.Printf("[INFO] Found %d childnodes for action %s (%s)", len(childNodes), action.Label, action.ID)
		}
	*/

	// FIXME - append all nodes (actions, triggers etc) to one single array here
	//log.Printf("PRE VARIABLES")
	if len(foundNodes) != len(allNodes) || len(workflow.Actions) <= 0 {
		// This shit takes a few seconds lol
		if !workflow.IsValid {
			oldworkflow, err := GetWorkflow(ctx, fileId)
			if err != nil {
				log.Printf("[WARNING] Workflow %s doesn't exist - oldworkflow.", fileId)
				if workflow.PreviouslySaved {
					resp.WriteHeader(401)
					resp.Write([]byte(`{"success": false, "reason": "Item already exists."}`))
					return
				}
			}

			oldworkflow.IsValid = false
			err = SetWorkflow(ctx, *oldworkflow, fileId)
			if err != nil {
				log.Printf("[WARNING] Failed saving workflow to database: %s", err)
				if workflow.PreviouslySaved {
					resp.WriteHeader(401)
					resp.Write([]byte(`{"success": false}`))
					return
				}
			}

			cacheKey := fmt.Sprintf("%s_workflows", user.Id)
			DeleteCache(ctx, cacheKey)
		}
	}

	// FIXME - might be a sploit to run someone elses app if getAllWorkflowApps
	// doesn't check sharing=true
	// Have to do it like this to add the user's apps
	//log.Printf("EXIT ON ERROR: %s", workflow.Configuration.ExitOnError)

	// Started getting the single apps, but if it's weird, this is faster
	// 1. Check workflow.Start
	// 2. Check if any node has "isStartnode"
	//if len(workflow.Actions) > 0 {
	//	index := -1
	//	for indexFound, action := range workflow.Actions {
	//		if workflow.Start == action.ID {
	//			index = indexFound
	//		}
	//	}

	//	if index >= 0 {
	//		workflow.Actions[0].IsStartNode = true
	//	} else {
	//		log.Printf("[WARNING] Couldn't find startnode %s!", workflow.Start)
	//		if workflow.PreviouslySaved {
	//			resp.WriteHeader(401)
	//			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "You need to set a startnode."}`)))
	//			return
	//		}
	//	}
	//}

	/*
		allAuths, err := GetAllWorkflowAppAuth(ctx, user.ActiveOrg.Id)
		if userErr != nil {
			log.Printf("Api authentication failed in get all apps: %s", userErr)
			if workflow.PreviouslySaved {
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false}`))
				return
			}
		}
	*/

	// Check every app action and param to see whether they exist
	//log.Printf("PRE ACTIONS 2")
	allAuths, autherr := GetAllWorkflowAppAuth(ctx, user.ActiveOrg.Id)
	newActions = []Action{}
	for _, action := range workflow.Actions {
		reservedApps := []string{
			"0ca8887e-b4af-4e3e-887c-87e9d3bc3d3e",
		}

		//log.Printf("%s Action execution var: %s", action.Label, action.ExecutionVariable.Name)

		builtin := false
		for _, id := range reservedApps {
			if id == action.AppID {
				builtin = true
				break
			}
		}

		// Check auth
		// 1. Find the auth in question
		// 2. Update the node and workflow info in the auth
		// 3. Get the values in the auth and add them to the action values
		handleOauth := false
		if len(action.AuthenticationId) > 0 {
			//log.Printf("\n\nLen: %d", len(allAuths))
			authFound := false
			for _, auth := range allAuths {
				if auth.Id == action.AuthenticationId {
					authFound = true

					if strings.ToLower(auth.Type) == "oauth2" {
						handleOauth = true
					}

					// Updates the auth item itself IF necessary
					UpdateAppAuth(ctx, auth, workflow.ID, action.ID, true)
					break
				}
			}

			if !authFound {
				log.Printf("[WARNING] App auth %s doesn't exist. Setting error", action.AuthenticationId)

				errorMsg := fmt.Sprintf("App authentication %s for app %s doesn't exist!", action.AuthenticationId, action.AppName)
				if !ArrayContains(workflow.Errors, errorMsg) {
					workflow.Errors = append(workflow.Errors, errorMsg)
				}
				workflow.IsValid = false

				action.Errors = append(action.Errors, "App authentication doesn't exist")
				action.IsValid = false
				action.AuthenticationId = ""
				//resp.WriteHeader(401)
				//resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "App auth %s doesn't exist"}`, action.AuthenticationId)))
				//return
			}
		}

		if builtin {
			newActions = append(newActions, action)
		} else {
			curapp := WorkflowApp{}

			// ID first, then name + version
			// If it can't find param, it will swap it over farther down
			for _, app := range workflowapps {
				if app.ID == action.AppID {
					curapp = app
					break
				}
			}

			if curapp.ID == "" {
				//log.Printf("[WARNING] Didn't find the App ID for %s", action.AppID)
				for _, app := range workflowapps {
					if app.ID == action.AppID {
						curapp = app
						break
					}

					// Has to NOT be generated
					if app.Name == action.AppName {
						if app.AppVersion == action.AppVersion {
							curapp = app
							break
						} else if ArrayContains(app.LoopVersions, action.AppVersion) {
							// Get the real app
							for _, item := range app.Versions {
								if item.Version == action.AppVersion {
									//log.Printf("Should get app %s - %s", item.Version, item.ID)

									tmpApp, err := GetApp(ctx, item.ID, user, false)
									if err != nil && tmpApp.ID == "" {
										log.Printf("[WARNING] Failed getting app %s (%s): %s", app.Name, item.ID, err)
									}

									curapp = *tmpApp
									break
								}
							}

							//curapp = app
							break
						}
					}
				}
			} else {
				//log.Printf("[DEBUG] Found correct App ID for %s", action.AppID)
			}

			//log.Printf("CURAPP: %s:%s", curapp.Name, curapp.AppVersion)

			if curapp.Name != action.AppName {
				errorMsg := fmt.Sprintf("App %s:%s doesn't exist", action.AppName, action.AppVersion)
				action.Errors = append(action.Errors, "This app doesn't exist.")

				if !ArrayContains(workflow.Errors, errorMsg) {
					workflow.Errors = append(workflow.Errors, errorMsg)
					log.Printf("[WARNING] App %s:%s doesn't exist. Adding as error.", action.AppName, action.AppVersion)
				}

				action.IsValid = false
				workflow.IsValid = false

				// Append with errors
				newActions = append(newActions, action)
				//resp.WriteHeader(401)
				//resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "App %s doesn't exist"}`, action.AppName)))
				//return
			} else {
				// Check to see if the appaction is valid
				curappaction := WorkflowAppAction{}
				for _, curAction := range curapp.Actions {
					if action.Name == curAction.Name {
						curappaction = curAction
						break
					}
				}

				if curappaction.Name != action.Name {
					// FIXME: Check if another app with the same name has the action here
					// Then update the ID? May be related to updated apps etc.
					//log.Printf("Couldn't find action - checking similar apps")
					for _, app := range workflowapps {
						if app.ID == curapp.ID {
							continue
						}

						// Has to NOT be generated
						if app.Name == action.AppName && app.AppVersion == action.AppVersion {
							for _, curAction := range app.Actions {
								if action.Name == curAction.Name {
									log.Printf("[DEBUG] Found app %s (NOT %s) with the param: %s", app.ID, curapp.ID, curAction.Name)
									curappaction = curAction
									action.AppID = app.ID
									curapp = app
									break
								}
							}
						}

						if curappaction.Name == action.Name {
							break
						}
					}
				}

				// Check to see if the action is valid
				if curappaction.Name != action.Name {
					// FIXME: Find the actual active app?

					log.Printf("[ERROR] Action '%s' in app %s doesn't exist. Workflow: %s (%s)", action.Name, curapp.Name, workflow.Name, workflow.ID)
					thisError := fmt.Sprintf("%s: Action %s in app %s doesn't exist", action.Label, action.Name, action.AppName)
					workflow.Errors = append(workflow.Errors, thisError)
					workflow.IsValid = false

					if !ArrayContains(action.Errors, thisError) {
						action.Errors = append(action.Errors, thisError)
					}

					action.IsValid = false
					//if workflow.PreviouslySaved {
					//	resp.WriteHeader(401)
					//	resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Action %s in app %s doesn't exist"}`, action.Name, curapp.Name)))
					//	return
					//}
				}

				// FIXME - check all parameters to see if they're valid
				// Includes checking required fields

				selectedAuth := AppAuthenticationStorage{}
				if len(action.AuthenticationId) > 0 && autherr == nil {
					for _, auth := range allAuths {
						if auth.Id == action.AuthenticationId {
							selectedAuth = auth
							break
						}
					}
				}

				newParams := []WorkflowAppActionParameter{}
				for _, param := range curappaction.Parameters {
					paramFound := false

					// Handles check for parameter exists + value not empty in used fields
					for _, actionParam := range action.Parameters {
						if actionParam.Name == param.Name {
							paramFound = true

							if actionParam.Value == "" && actionParam.Variant == "STATIC_VALUE" && actionParam.Required == true {
								// Validating if the field is an authentication field
								if len(selectedAuth.Id) > 0 {
									authFound := false
									for _, field := range selectedAuth.Fields {
										if field.Key == actionParam.Name {
											authFound = true
											//log.Printf("FOUND REQUIRED KEY %s IN AUTH", field.Key)
											break
										}
									}

									if authFound {
										newParams = append(newParams, actionParam)
										continue
									}
								}

								//log.Printf("[WARNING] Appaction %s with required param '%s' is empty. Can't save.", action.Name, param.Name)
								thisError := fmt.Sprintf("%s is missing required parameter %s", action.Label, param.Name)
								if handleOauth {
									//log.Printf("[WARNING] Handling oauth2 app saving, hence not throwing warnings (1)")
									//workflow.Errors = append(workflow.Errors, fmt.Sprintf("Debug: Handling one Oauth2 app (%s). May cause issues during initial configuration (1)", action.Name))
								} else {
									action.Errors = append(action.Errors, thisError)
									workflow.Errors = append(workflow.Errors, thisError)
									action.IsValid = false
								}
							}

							if actionParam.Variant == "" {
								actionParam.Variant = "STATIC_VALUE"
							}

							newParams = append(newParams, actionParam)
							break
						}
					}

					// Handles check for required params
					if !paramFound && param.Required {
						if handleOauth {
							log.Printf("[WARNING] Handling oauth2 app saving, hence not throwing warnings (2)")
							//workflow.Errors = append(workflow.Errors, fmt.Sprintf("Debug: Handling one Oauth2 app (%s). May cause issues during initial configuration (2)", action.Name))
						} else {
							thisError := fmt.Sprintf("Parameter %s is required", param.Name)
							action.Errors = append(action.Errors, thisError)

							workflow.Errors = append(workflow.Errors, thisError)
							action.IsValid = false
						}

						//newActions = append(newActions, action)
						//resp.WriteHeader(401)
						//resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Appaction %s with required param '%s' is empty."}`, action.Name, param.Name)))
						//return
					}

				}

				action.Parameters = newParams
				newActions = append(newActions, action)
			}
		}
	}

	if !workflow.PreviouslySaved {
		log.Printf("[WORKFLOW INIT] NOT PREVIOUSLY SAVED - SET ACTION AUTH!")

		if autherr == nil && len(workflowapps) > 0 && apperr == nil {
			//log.Printf("Setting actions")
			actionFixing := []Action{}
			appsAdded := []string{}
			for _, action := range newActions {
				setAuthentication := false
				if len(action.AuthenticationId) > 0 {
					//found := false
					authenticationFound := false
					for _, auth := range allAuths {
						if auth.Id == action.AuthenticationId {
							authenticationFound = true
							break
						}
					}

					if !authenticationFound {
						setAuthentication = true
					}
				} else {
					// FIXME: 1. Validate if the app needs auth
					// 1. Validate if auth for the app exists
					// var appAuth AppAuthenticationStorage
					setAuthentication = true

					//App           WorkflowApp           `json:"app" datastore:"app,noindex"`
				}

				if setAuthentication {
					authSet := false
					for _, auth := range allAuths {
						if !auth.Active {
							continue
						}

						if !auth.Defined {
							continue
						}

						if auth.App.Name == action.AppName {
							//log.Printf("FOUND AUTH FOR APP %s: %s", auth.App.Name, auth.Id)
							action.AuthenticationId = auth.Id
							authSet = true
							break
						}
					}

					// FIXME: Only o this IF there isn't another one for the app already
					if !authSet {
						//log.Printf("Validate if the app NEEDS auth or not")
						outerapp := WorkflowApp{}
						for _, app := range workflowapps {
							if app.Name == action.AppName {
								outerapp = app
								break
							}
						}

						if len(outerapp.ID) > 0 && outerapp.Authentication.Required {
							found := false
							for _, auth := range allAuths {
								if auth.App.ID == outerapp.ID {
									found = true
									break
								}
							}

							for _, added := range appsAdded {
								if outerapp.ID == added {
									found = true
								}
							}

							// FIXME: Add app auth
							if !found {
								timeNow := int64(time.Now().Unix())
								authFields := []AuthenticationStore{}
								for _, param := range outerapp.Authentication.Parameters {
									authFields = append(authFields, AuthenticationStore{
										Key:   param.Name,
										Value: "",
									})
								}

								appAuth := AppAuthenticationStorage{
									Active:        true,
									Label:         fmt.Sprintf("default_%s", outerapp.Name),
									Id:            uuid.NewV4().String(),
									App:           outerapp,
									Fields:        authFields,
									Usage:         []AuthenticationUsage{},
									WorkflowCount: 0,
									NodeCount:     0,
									OrgId:         user.ActiveOrg.Id,
									Created:       timeNow,
									Edited:        timeNow,
								}

								err = SetWorkflowAppAuthDatastore(ctx, appAuth, appAuth.Id)
								if err != nil {
									log.Printf("Failed setting appauth for with name %s", appAuth.Label)
								} else {
									appsAdded = append(appsAdded, outerapp.ID)
								}
							}

							action.Errors = append(action.Errors, "Requires authentication")
							action.IsValid = false
							workflow.IsValid = false
						}
					}
				}

				actionFixing = append(actionFixing, action)
			}

			newActions = actionFixing
		} else {
			log.Printf("FirstSave error: %s - %s", err, apperr)
			//allAuths, err := GetAllWorkflowAppAuth(ctx, user.ActiveOrg.Id)
		}

		skipSave, skipSaveOk := request.URL.Query()["skip_save"]
		if skipSaveOk && len(skipSave) > 0 {
			//log.Printf("INSIDE SKIPSAVE: %s", skipSave[0])
			if strings.ToLower(skipSave[0]) != "true" {
				workflow.PreviouslySaved = true
			}
		} else {
			workflow.PreviouslySaved = true
		}
	}
	//log.Printf("SAVED: %s", workflow.PreviouslySaved)

	workflow.Actions = newActions
	workflow.IsValid = true

	// TBD: Is this too drastic? May lead to issues in the future.
	if workflow.OrgId != user.ActiveOrg.Id {
		log.Printf("[WARNING] Editing workflow to be owned by org %s", user.ActiveOrg.Id)
		workflow.OrgId = user.ActiveOrg.Id
		workflow.ExecutingOrg = user.ActiveOrg
		workflow.Org = append(workflow.Org, user.ActiveOrg)
	}

	// Only happens if the workflow is public and being edited
	if correctUser {
		workflow.Public = true

		// Should save it in Algolia too?
		_, err = handleAlgoliaWorkflowUpdate(ctx, workflow)
		if err != nil {
			log.Printf("[ERROR] Failed finding publicly changed workflow %s for user %s (%s): %s", workflow.ID, user.Username, user.Id, err)
		} else {
			log.Printf("[DEBUG] User %s (%s) updated their public workflow %s (%s)", user.Username, user.Id, workflow.Name, workflow.ID)
		}
	}

	err = SetWorkflow(ctx, workflow, fileId)
	if err != nil {
		log.Printf("[WARNING] Failed saving workflow to database: %s", err)
		if workflow.PreviouslySaved {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	if org.Id == "" {
		org, err = GetOrg(ctx, user.ActiveOrg.Id)
		if err != nil {
			log.Printf("[WARNING] Failed getting org during wf save of %s (org: %s): %s", workflow.ID, user.ActiveOrg.Id, err)
		}
	}

	// This may cause some issues with random slow loads with cross & suborgs, but that's fine (for now)
	// FIX: Should only happen for users with this org as the active one
	// Org-based workflows may also work
	if org.Id != "" {
		for _, loopUser := range org.Users {
			DeleteCache(ctx, fmt.Sprintf("%s_workflows", loopUser.Id))
		}
	}

	if orgUpdated {
		err = SetOrg(ctx, *org, org.Id)
		if err != nil {
			log.Printf("[WARNING] Failed setting org when autoadding apps and updating framework on save workflow save (%s): %s", workflow.ID, err)
		} else {
			log.Printf("[DEBUG] Successfully updated org %s during save of %s for user %s (%s", user.ActiveOrg.Id, workflow.ID, user.Username, user.Id)
		}
	}

	//totalOldActions := len(tmpworkflow.Actions)
	//totalNewActions := len(workflow.Actions)
	//err = increaseStatisticsField(ctx, "total_workflow_actions", workflow.ID, int64(totalNewActions-totalOldActions), workflow.OrgId)
	//if err != nil {
	//	log.Printf("Failed to change total actions data: %s", err)
	//}

	type returnData struct {
		Success bool     `json:"success"`
		Errors  []string `json:"errors"`
	}

	returndata := returnData{
		Success: true,
		Errors:  workflow.Errors,
	}

	// Really don't know why this was happening
	//cacheKey := fmt.Sprintf("workflowapps-sorted-100")
	//requestCache.Delete(cacheKey)
	//cacheKey = fmt.Sprintf("workflowapps-sorted-500")
	//requestCache.Delete(cacheKey)

	log.Printf("[INFO] Saved new version of workflow %s (%s) for org %s. Actions: %d, Triggers: %d", workflow.Name, fileId, workflow.OrgId, len(workflow.Actions), len(workflow.Triggers))
	resp.WriteHeader(200)
	newBody, err := json.Marshal(returndata)
	if err != nil {
		resp.Write([]byte(`{"success": true}`))
		return
	}

	resp.Write(newBody)
}

func HandleCategoryIncrease(categories Categories, action Action, workflowapps []WorkflowApp) Categories {
	if action.Category == "" {
		appName := action.AppName
		for _, app := range workflowapps {
			if appName != strings.ToLower(app.Name) {
				continue
			}

			if len(app.Categories) > 0 {
				log.Printf("[INFO] Setting category for %s: %s", app.Name, app.Categories)
				action.Category = app.Categories[0]
				break
			}
		}

		//log.Printf("Should find app's categories as it's empty during save")
		return categories
	}

	//log.Printf("Action: %s, category: %s", action.AppName, action.Category)
	// FIXME: Make this an "autodiscover" that's controlled by the category itself
	// Should just be a list that's looped against :)
	newCategory := strings.ToLower(action.Category)
	if strings.Contains(newCategory, "case") || strings.Contains(newCategory, "ticket") || strings.Contains(newCategory, "alert") || strings.Contains(newCategory, "mssp") {
		categories.Cases.Count += 1
	} else if strings.Contains(newCategory, "siem") || strings.Contains(newCategory, "event") || strings.Contains(newCategory, "log") || strings.Contains(newCategory, "search") {
		categories.SIEM.Count += 1
	} else if strings.Contains(newCategory, "sms") || strings.Contains(newCategory, "comm") || strings.Contains(newCategory, "phone") || strings.Contains(newCategory, "call") || strings.Contains(newCategory, "chat") || strings.Contains(newCategory, "mail") || strings.Contains(newCategory, "phish") {
		categories.Communication.Count += 1
	} else if strings.Contains(newCategory, "intel") || strings.Contains(newCategory, "crim") || strings.Contains(newCategory, "ti") {
		categories.Intel.Count += 1
	} else if strings.Contains(newCategory, "sand") || strings.Contains(newCategory, "virus") || strings.Contains(newCategory, "malware") || strings.Contains(newCategory, "scan") || strings.Contains(newCategory, "edr") || strings.Contains(newCategory, "endpoint detection") {
		// Sandbox lol
		categories.EDR.Count += 1
	} else if strings.Contains(newCategory, "vuln") || strings.Contains(newCategory, "fim") || strings.Contains(newCategory, "fim") || strings.Contains(newCategory, "integrity") {
		categories.Assets.Count += 1
	} else if strings.Contains(newCategory, "network") || strings.Contains(newCategory, "firewall") || strings.Contains(newCategory, "waf") || strings.Contains(newCategory, "switch") {
		categories.Network.Count += 1
	} else {
		categories.Other.Count += 1
	}

	return categories
}

// Adds app auth tracking
func UpdateAppAuth(ctx context.Context, auth AppAuthenticationStorage, workflowId, nodeId string, add bool) error {
	workflowFound := false
	workflowIndex := 0
	nodeFound := false
	for index, workflow := range auth.Usage {
		if workflow.WorkflowId == workflowId {
			// Check if node exists
			workflowFound = true
			workflowIndex = index
			for _, actionId := range workflow.Nodes {
				if actionId == nodeId {
					nodeFound = true
					break
				}
			}

			break
		}
	}

	// FIXME: Add a way to use !add to remove
	updateAuth := false
	if !workflowFound && add {
		//log.Printf("[INFO] Adding workflow things to auth!")
		usageItem := AuthenticationUsage{
			WorkflowId: workflowId,
			Nodes:      []string{nodeId},
		}

		auth.Usage = append(auth.Usage, usageItem)
		auth.WorkflowCount += 1
		auth.NodeCount += 1
		updateAuth = true
	} else if !nodeFound && add {
		//log.Printf("[INFO] Adding node things to auth!")
		auth.Usage[workflowIndex].Nodes = append(auth.Usage[workflowIndex].Nodes, nodeId)
		auth.NodeCount += 1
		updateAuth = true
	}

	if updateAuth {
		//log.Printf("[INFO] Updating auth!")
		err := SetWorkflowAppAuthDatastore(ctx, auth, auth.Id)
		if err != nil {
			log.Printf("[WARNING] Failed UPDATING app auth %s: %s", auth.Id, err)
			return err
		}
	}

	return nil
}

func HandleApiGeneration(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting API GEN request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	userInfo, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in apigen: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	//log.Printf("IN APIKEY GENERATION")
	ctx := GetContext(request)
	if request.Method == "GET" {
		newUserInfo, err := GenerateApikey(ctx, userInfo)
		if err != nil {
			log.Printf("[WARNING] Failed to generate apikey for user %s: %s", userInfo.Username, err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": ""}`))
			return
		}

		userInfo = newUserInfo
		log.Printf("[INFO] Updated apikey for user %s", userInfo.Username)
	} else if request.Method == "POST" {
		if request.Body == nil {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Printf("Failed reading body")
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Missing field: user_id"}`)))
			return
		}

		type userId struct {
			UserId string `json:"user_id"`
		}

		var t userId
		err = json.Unmarshal(body, &t)
		if err != nil {
			log.Printf("Failed unmarshaling userId: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unmarshaling. Missing field: user_id"}`)))
			return
		}

		log.Printf("[INFO] Handling post for APIKEY gen FROM user %s. Userchange: %s!", userInfo.Username, t.UserId)

		if userInfo.Role != "admin" {
			log.Printf("[AUDIT] %s tried and failed to change apikey for %s (2)", userInfo.Username, t.UserId)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "You need to be admin to change others' apikey"}`)))
			return
		}

		foundUser, err := GetUser(ctx, t.UserId)
		if err != nil {
			log.Printf("[INFO] Can't find user %s (apikey gen): %s", t.UserId, err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false}`)))
			return
		}

		// FIXME: May not be good due to different roles in different organizations.
		if foundUser.Role == "admin" {
			log.Printf("[AUDIT] %s tried and failed to change apikey for %s. Skipping because users' role is admin", userInfo.Username, t.UserId)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can't change the apikey of another admin"}`)))
			return
		}

		newUserInfo, err := GenerateApikey(ctx, *foundUser)
		if err != nil {
			log.Printf("Failed to generate apikey for user %s: %s", foundUser.Username, err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
			return
		}

		foundUser = &newUserInfo

		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf(`{"success": true, "username": "%s", "verified": %t, "apikey": "%s"}`, foundUser.Username, foundUser.Verified, foundUser.ApiKey)))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true, "username": "%s", "verified": %t, "apikey": "%s"}`, userInfo.Username, userInfo.Verified, userInfo.ApiKey)))
}

func HandleSettings(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting Handle Settings request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	userInfo, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in settings: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	newObject := SettingsReturn{
		Success:  true,
		Username: userInfo.Username,
		Verified: userInfo.Verified,
		Apikey:   userInfo.ApiKey,
		Image:    userInfo.PublicProfile.GithubAvatar,
	}

	newjson, err := json.Marshal(newObject)
	if err != nil {
		log.Printf("[ERROR] Failed unmarshal in get settings: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed handling your user"}`)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func HandleGetUsers(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting Get Users request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get users: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role != "admin" {
		log.Printf("[AUDIT] User isn't admin (%s) and can't list users.", user.Role)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Not admin"}`))
		return
	}

	ctx := GetContext(request)
	org, err := GetOrg(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Failed getting org in get users: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed getting org when listing users"}`))
		return
	}

	newUsers := []User{}
	for _, item := range org.Users {
		if len(item.Username) == 0 {
			continue
		}

		//for _, tmpUser := range newUsers {
		//	if tmpUser.Name
		//}

		if item.Username != user.Username && (len(item.Orgs) > 1 || item.Role == "admin") {
			//log.Printf("[DEBUG] Orgs for the user: %s", item.Orgs)
			item.ApiKey = ""
		}

		item.Password = ""
		item.Session = ""
		item.VerificationToken = ""
		item.Orgs = []string{}
		item.EthInfo = EthInfo{}

		// Will get from cache 2nd time so this is fine.
		if user.Id == item.Id {
			item.Orgs = user.Orgs
			item.Active = user.Active
			item.MFA = user.MFA
		} else {
			foundUser, err := GetUser(ctx, item.Id)
			if err == nil {
				// Only add IF the admin querying it has access, meaning only show what you yourself have access toMFAInfo
				allOrgs := []string{}
				for _, orgname := range foundUser.Orgs {
					found := false

					for _, userOrg := range user.Orgs {
						if userOrg == orgname {
							found = true
							break
						}
					}

					if found {
						allOrgs = append(allOrgs, orgname)
					}
				}

				//log.Printf("[DEBUG] Added %d org(s) for user %s (%s) - get users", len(allOrgs), foundUser.Username, foundUser.Id)

				item.MFA = foundUser.MFA
				item.Verified = foundUser.Verified
				item.Active = foundUser.Active
				item.Orgs = allOrgs
			}
		}

		if len(item.Orgs) == 0 {
			item.Orgs = append(item.Orgs, user.ActiveOrg.Id)
		}

		newUsers = append(newUsers, item)
	}

	deduplicatedUsers := []User{}
	for _, item := range newUsers {
		found := false
		for _, tmpUser := range deduplicatedUsers {
			if tmpUser.Username == item.Username {
				found = true
				break
			}
		}

		if !found {
			//log.Printf("[DEBUG] Adding user %s (%s) to list", item.Username, item.Id)
			deduplicatedUsers = append(deduplicatedUsers, item)
		}
	}

	newjson, err := json.Marshal(deduplicatedUsers)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshal in getusers: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking"}`)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func HandlePasswordChange(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting Password Change request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	if request.Body == nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Failed reading body")
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false}`)))
		return
	}

	// Get the current user - check if they're admin or the "username" user.
	var t PasswordChange
	err = json.Unmarshal(body, &t)
	if err != nil {
		log.Printf("Failed unmarshaling")
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false}`)))
		return
	}

	userInfo, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in password change: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	log.Printf("[AUDIT] Handling password change for %s from %s (%s)", t.Username, userInfo.Username, userInfo.Id)

	curUserFound := false
	if t.Username != userInfo.Username {
		log.Printf("[WARNING] Bad username during password change for %s.", t.Username)

		if project.Environment == "cloud" {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Not allowed to change others' passwords in cloud"}`))
			return
		}
	} else if t.Username == userInfo.Username {
		curUserFound = true
	}

	if userInfo.Role != "admin" {
		if t.Newpassword != t.Newpassword2 {
			err := "Passwords don't match"
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
			return
		}

		if project.Environment == "cloud" {
			if len(t.Newpassword) < 10 || len(t.Newpassword2) < 10 {
				err := "Passwords too short - 2"
				resp.WriteHeader(401)
				resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
				return
			}
		}
	} else {
		// Check ORG HERE?
	}

	// Current password
	err = CheckPasswordStrength(t.Newpassword)
	if err != nil {
		log.Printf("[INFO] Bad password strength: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
		return
	}

	ctx := GetContext(request)
	foundUser := User{}
	if !curUserFound {
		users, err := FindUser(ctx, strings.ToLower(strings.TrimSpace(t.Username)))
		if err != nil && len(users) == 0 {
			log.Printf("[WARNING] Failed getting user %s: %s", t.Username, err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Username and/or password is incorrect"}`))
			return
		}

		if len(users) != 1 {
			log.Printf(`[WARNING] Found multiple or no users with the same username: %s: %d`, t.Username, len(users))
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Found %d users with the same username: %s"}`, len(users), t.Username)))
			return
		}

		foundUser = users[0]
		orgFound := false
		if userInfo.ActiveOrg.Id == foundUser.ActiveOrg.Id {
			orgFound = true
		} else {
			for _, item := range foundUser.Orgs {
				if item == userInfo.ActiveOrg.Id {
					orgFound = true
					break
				}
			}
		}

		if !orgFound {
			log.Printf("[AUDIT] User %s (%s) is admin, but can't change user's (%s) password outside their own org.", userInfo.Username, userInfo.Id, foundUser.Username)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can't change users outside your org."}`)))
			return
		}
	} else {
		// Admins can re-generate others' passwords as well.
		if userInfo.Role != "admin" {
			err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(t.Currentpassword))
			if err != nil {
				log.Printf("[WARNING] Bad password for %s: %s", userInfo.Username, err)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false, "reason": "Username and/or password is incorrect"}`))
				return
			}
		}

		// log.Printf("FOUND: %s", curUserFound)
		foundUser = userInfo
		//userInfo, err := HandleApiAuthentication(resp, request)
	}

	if len(foundUser.Id) == 0 {
		log.Printf("[WARNING] Something went wrong in password reset: couldn't find user.")
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(t.Newpassword), 8)
	if err != nil {
		log.Printf("New password failure for %s: %s", userInfo.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Username and/or password is incorrect"}`))
		return
	}

	foundUser.Password = string(hashedPassword)
	err = SetUser(ctx, &foundUser, true)
	if err != nil {
		log.Printf("Error fixing password for user %s: %s", userInfo.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Username and/or password is incorrect"}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true}`)))
}

// Can check against HIBP etc?
// Removed for localhost
func CheckPasswordStrength(password string) error {
	// Check password strength here
	if len(password) < 4 {
		return errors.New("Minimum password length is 4.")
	}

	//if len(password) > 128 {
	//	return errors.New("Maximum password length is 128.")
	//}

	//re := regexp.MustCompile("[0-9]+")
	//if len(re.FindAllString(password, -1)) == 0 {
	//	return errors.New("Password must contain a number")
	//}

	//re = regexp.MustCompile("[a-z]+")
	//if len(re.FindAllString(password, -1)) == 0 {
	//	return errors.New("Password must contain a lower case char")
	//}

	//re = regexp.MustCompile("[A-Z]+")
	//if len(re.FindAllString(password, -1)) == 0 {
	//	return errors.New("Password must contain an upper case char")
	//}

	return nil
}

func SendHookResult(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in send hook results: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}
	_ = user

	location := strings.Split(request.URL.String(), "/")

	var workflowId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		workflowId = location[4]
	}

	if len(workflowId) != 32 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "message": "ID not valid"}`))
		return
	}

	ctx := GetContext(request)
	hook, err := GetHook(ctx, workflowId)
	if err != nil {
		log.Printf("[WARNING] Failed getting hook %s (send): %s", workflowId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Body data error: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	log.Printf("SET the hook results for %s to %s", workflowId, body)
	// FIXME - set the hook result in the DB somehow as interface{}
	// FIXME - should the hook do the transform? Hmm

	b, err := json.Marshal(hook)
	if err != nil {
		log.Printf("Failed marshalling: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte(b))
	return
}

func HandleGetHook(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get hook: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")

	var workflowId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		workflowId = location[4]
	}

	if len(workflowId) != 36 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "message": "ID not valid"}`))
		return
	}

	ctx := GetContext(request)
	hook, err := GetHook(ctx, workflowId)
	if err != nil {
		log.Printf("[WARNING] Failed getting hook %s (get hook): %s", workflowId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Id != hook.Owner && user.Role != "scheduler" {
		log.Printf("Wrong user (%s) for hook %s", user.Username, hook.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	b, err := json.Marshal(hook)
	if err != nil {
		log.Printf("Failed marshalling: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte(b))
	return
}

func GetSpecificWorkflow(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Removed check here as it may be a public workflow
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[AUDIT] Api authentication failed in getting specific workflow: %s. Continuing because it may be public.", err)
	}

	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if strings.Contains(fileId, "?") {
		fileId = strings.Split(fileId, "?")[0]
	}

	if len(fileId) != 36 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow ID when getting workflow is not valid"}`))
		return
	}

	ctx := GetContext(request)
	workflow, err := GetWorkflow(ctx, fileId)
	if err != nil {
		log.Printf("[WARNING] Workflow %s doesn't exist.", fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed finding workflow"}`))
		return
	}

	// Check workflow.Sharing == private / public / org  too
	if user.Id != workflow.Owner || len(user.Id) == 0 {
		// Added org-reader as the user should be able to read everything in an org
		//if workflow.OrgId == user.ActiveOrg.Id && (user.Role == "admin" || user.Role == "org-reader") {
		if workflow.OrgId == user.ActiveOrg.Id {
			log.Printf("[AUDIT] User %s is accessing workflow %s as admin (get workflow)", user.Username, workflow.ID)
		} else if workflow.Public {
			log.Printf("[AUDIT] Letting user %s access workflow %s because it's public", user.Username, workflow.ID)

			// Only for Read-Only. No executions or impersonations.
		} else if project.Environment == "cloud" && user.Verified == true && user.Active == true && user.SupportAccess == true && strings.HasSuffix(user.Username, "@shuffler.io") {
			log.Printf("[AUDIT] Letting verified support admin %s access workflow %s", user.Username, workflow.ID)
		} else {
			log.Printf("[AUDIT] Wrong user (%s) for workflow %s (get workflow). Verified: %t, Active: %t, SupportAccess: %t, Username: %s", user.Username, workflow.ID, user.Verified, user.Active, user.SupportAccess, user.Username)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	if len(workflow.Actions) == 0 {
		workflow.Actions = []Action{}
	}
	if len(workflow.Branches) == 0 {
		workflow.Branches = []Branch{}
	}
	if len(workflow.Triggers) == 0 {
		workflow.Triggers = []Trigger{}
	}
	if len(workflow.Errors) == 0 {
		workflow.Errors = []string{}
	}

	for key, _ := range workflow.Actions {
		// RefUrl not necessary anymore, as we migrated to getting apps during exec
		workflow.Actions[key].ReferenceUrl = ""

		// Never helpful when this is red
		if workflow.Actions[key].AppName == "Shuffle Tools" {
			workflow.Actions[key].IsValid = true
		}

		if !workflow.Actions[key].IsValid {
			log.Printf("[AUDIT] Invalid action: %s (%s)", workflow.Actions[key].Label, workflow.Actions[key].ID)

			// Check if all fields are set
			// Check if auth is set (autofilled)
			isValid := true
			for _, param := range workflow.Actions[key].Parameters {
				if param.Required && len(param.Value) == 0 {
					isValid = false
					break
				}
			}

			if isValid {
				workflow.Actions[key].IsValid = true
			}
		}
	}

	body, err := json.Marshal(workflow)
	if err != nil {
		log.Printf("[WARNING] Failed workflow GET marshalling: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write(body)
}

func DeleteUser(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting User request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	userInfo, userErr := HandleApiAuthentication(resp, request)
	if userErr != nil {
		log.Printf("[WARNING] Api authentication failed in delete user: %s", userErr)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if userInfo.Role != "admin" {
		log.Printf("Wrong user (%s) when deleting - must be admin", userInfo.Username)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Must be admin"}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")
	var userId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		userId = location[4]
	}

	ctx := GetContext(request)
	userId, err := url.QueryUnescape(userId)
	if err != nil {
		log.Printf("[WARNING] Failed decoding user %s: %s", userId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false}`)))
		return
	}

	if userId == userInfo.Id {
		log.Printf("[WARNING] Can't change activation of your own user %s (%s)", userInfo.Username, userInfo.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can't change activation of your own user"}`)))
		return
	}

	foundUser, err := GetUser(ctx, userId)
	if err != nil {
		log.Printf("[WARNING] Can't find user %s (delete user): %s", userId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false}`)))
		return
	}

	orgFound := false
	if userInfo.ActiveOrg.Id == foundUser.ActiveOrg.Id {
		orgFound = true
	} else {
		for _, item := range foundUser.Orgs {
			if item == userInfo.ActiveOrg.Id {
				orgFound = true
				break
			}
		}
	}

	if !orgFound {
		log.Printf("[AUDIT] User %s is admin, but can't delete users outside their own org.", userInfo.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can't change users outside your org."}`)))
		return
	}

	// OLD: Invert. No user deletion.
	//if foundUser.Active {
	//	foundUser.Active = false
	//} else {
	//	foundUser.Active = true
	//}

	// NEW
	neworgs := []string{}
	for _, orgid := range foundUser.Orgs {
		if orgid == userInfo.ActiveOrg.Id {
			continue
		} else {
			// Automatically setting to first one
			if foundUser.ActiveOrg.Id == userInfo.ActiveOrg.Id {
				foundUser.ActiveOrg.Id = orgid
			}
		}

		neworgs = append(neworgs, orgid)
	}

	if foundUser.ActiveOrg.Id == userInfo.ActiveOrg.Id {
		log.Printf("[ERROR] User %s (%s) doesn't have an org anymore after being deleted. Give them one (NOT SET UP)", foundUser.Username, foundUser.Id)
		foundUser.ActiveOrg.Id = ""
	}

	foundUser.Orgs = neworgs
	err = SetUser(ctx, foundUser, false)
	if err != nil {
		log.Printf("[WARNING] Failed removing user %s (%s) from org %s", foundUser.Username, foundUser.Id, orgFound)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false}`)))
		return
	}

	org, err := GetOrg(ctx, userInfo.ActiveOrg.Id)
	if err != nil {
		log.Printf("[DEBUG] Failed getting org %s in delete user: %s", userInfo.ActiveOrg.Id, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false}`)))
		return
	}

	users := []User{}
	for _, user := range org.Users {
		if user.Id == foundUser.Id {
			continue
		}

		users = append(users, user)
	}

	org.Users = users
	err = SetOrg(ctx, *org, org.Id)
	if err != nil {
		log.Printf("[WARNING] Failed updating org (delete user %s) %s: %s", foundUser.Username, org.Id, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Removed their access but failed updating own user list"}`)))
		return
	}

	log.Printf("[INFO] Successfully removed %s from org %s", foundUser.Username, userInfo.ActiveOrg.Id)

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

// Only used for onprem :/
func UpdateWorkflowAppConfig(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting LOGIN request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	user, userErr := HandleApiAuthentication(resp, request)
	if userErr != nil {
		log.Printf("[AUDIT] Api authentication failed in get all apps: %s", userErr)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to edit apps")
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	ctx := GetContext(request)
	app, err := GetApp(ctx, fileId, user, false)
	if err != nil {
		log.Printf("[WARNING] Error getting app (update app): %s", fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Id != app.Owner {
		log.Printf("[WARNING] Wrong user (%s) for app %s in update app", user.Username, app.Name)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error with body read in update app: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Public means it's literally public to anyone right away.
	type updatefields struct {
		Sharing       bool   `json:"sharing"`
		SharingConfig string `json:"sharing_config"`
		Public        bool   `json:"public"`
	}

	var tmpfields updatefields
	err = json.Unmarshal(body, &tmpfields)
	if err != nil {
		log.Printf("[WARNING] Error with unmarshal body in update app: %s\n%s", err, string(body))
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if tmpfields.Sharing != app.Sharing {
		log.Printf("[INFO] Changing app sharing for %s to %s", app.ID, tmpfields.Sharing)
		app.Sharing = tmpfields.Sharing

		if project.Environment != "cloud" {
			log.Printf("[INFO] Set app %s (%s) to share everywhere (PUBLIC=true/false), because running onprem", app.Name, app.ID)
			app.Public = app.Sharing
		}
	}

	if tmpfields.SharingConfig != app.SharingConfig {
		log.Printf("[INFO] Changing app sharing CONFIG for %s to %s", app.ID, tmpfields.SharingConfig)
		app.SharingConfig = tmpfields.SharingConfig
	}

	if tmpfields.Public != app.Public {
		log.Printf("[INFO] Changing app %s to PUBLIC (THIS IS DEACTIVATED!)", app.ID)
		//app.Public = tmpfields.Public
	}

	err = SetWorkflowAppDatastore(ctx, *app, app.ID)
	if err != nil {
		log.Printf("[WARNING] Failed patching workflowapp: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	changed := false
	for index, privateApp := range user.PrivateApps {
		if privateApp.ID == app.ID {
			user.PrivateApps[index] = *app
			changed = true
			break
		}
	}

	if changed {
		err = SetUser(ctx, &user, true)
		if err != nil {
			log.Printf("[WARNING] Failed updating privateapp %s for user %s: %s", app.ID, user.Username, err)
		}
	}

	cacheKey := fmt.Sprintf("workflowapps-sorted-100")
	DeleteCache(ctx, cacheKey)
	cacheKey = fmt.Sprintf("workflowapps-sorted-500")
	DeleteCache(ctx, cacheKey)
	cacheKey = fmt.Sprintf("workflowapps-sorted-1000")
	DeleteCache(ctx, cacheKey)
	DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))

	log.Printf("[INFO] Changed App configuration for %s (%s)", app.Name, app.ID)
	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true}`)))
}

func deactivateApp(ctx context.Context, user User, app *WorkflowApp) error {
	//log.Printf("Should deactivate app %s\n for user %s", app, user)
	org, err := GetOrg(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[DEBUG] Failed getting org %s: %s", user.ActiveOrg.Id, err)
		return err
	}

	if !ArrayContains(org.ActiveApps, app.ID) {
		log.Printf("[WARNING] App %s isn't active for org %s", app.ID, user.ActiveOrg.Id)
		return errors.New(fmt.Sprintf("App %s isn't active for this org.", app.ID))
	}

	newApps := []string{}
	for _, appId := range org.ActiveApps {
		if appId == app.ID {
			continue
		}

		newApps = append(newApps, appId)
	}

	org.ActiveApps = newApps
	err = SetOrg(ctx, *org, org.Id)
	if err != nil {
		log.Printf("[WARNING] Failed updating org (deactive app %s) %s: %s", app.ID, org.Id, err)
		return err
	}

	return nil
}

func DeleteWorkflowApp(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, userErr := HandleApiAuthentication(resp, request)
	if userErr != nil {
		log.Printf("[WARNING] Api authentication failed in delete app: %s", userErr)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to delete apps: %s (%s)", user.Username, user.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	ctx := GetContext(request)
	app, err := GetApp(ctx, fileId, user, false)
	if err != nil {
		log.Printf("[WARNING] Error getting app %s: %s", app.Name, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Id != app.Owner {
		if user.Role == "admin" && app.Owner == "" {
			log.Printf("[INFO] Anyone can edit %s (%s), since it doesn't have an owner (DELETE).", app.Name, app.ID)
		} else {
			if user.Role == "admin" {
				err = deactivateApp(ctx, user, app)
				if err == nil {
					log.Printf("[INFO] App %s was deactivated for org %s", app.ID, user.ActiveOrg.Id)
					DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
					DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-100"))
					DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-500"))
					DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-1000"))
					DeleteCache(ctx, "all_apps")
					DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
					DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))
					resp.WriteHeader(200)
					resp.Write([]byte(`{"success": true}`))
					return
				}
			}

			log.Printf("[WARNING] Wrong user (%s) for app %s (%s) when DELETING app", user.Username, app.Name, app.ID)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	if (app.Public || app.Sharing) && project.Environment == "cloud" {
		log.Printf("[WARNING] App %s being deleted is public. Shouldn't be allowed. Public: %s, Sharing: %s", app.Name, app.Public, app.Sharing)

		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Can't delete public apps. Stop sharing it first, then delete it."}`))
		return
	}

	// Not really deleting it, just removing from user cache
	var privateApps []WorkflowApp
	for _, item := range user.PrivateApps {
		if item.ID == fileId {
			continue
		}

		privateApps = append(privateApps, item)
	}

	user.PrivateApps = privateApps

	err = SetUser(ctx, &user, true)
	if err != nil {
		log.Printf("[WARNING] Failed removing %s app for user %s: %s", app.Name, user.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false"}`)))
		return
	}

	err = DeleteKey(ctx, "workflowapp", app.ID)
	if err != nil {
		log.Printf("[WARNING] Failed deleting %s (%s) for by %s: %s", app.Name, app.ID, user.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false"}`)))
		return
	}

	// This is getting stupid :)
	DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-100"))
	DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-500"))
	DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-1000"))
	DeleteCache(ctx, "all_apps")
	DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
	DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
	DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

func HandleKeyValueCheck(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed reading body"}`))
		return
	}

	// Append: Checks if the value should be appended
	// WorkflowCheck: Checks if the value should only check for workflow in org, or entire org
	// Authorization: the Authorization to use
	// ExecutionRef: Ref for the execution
	// Values: The values to use
	type DataValues struct {
		App             string
		Actions         string
		ParameterNames  []string
		ParameterValues []string
	}

	type ReturnData struct {
		Append        bool         `json:"append"`
		WorkflowCheck bool         `json:"workflow_check"`
		Authorization string       `json:"authorization"`
		ExecutionRef  string       `json:"execution_ref"`
		OrgId         string       `json:"org_id"`
		Values        []DataValues `json:"values"`
	}

	//for key, value := range data.Apps {
	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	var tmpData ReturnData
	err = json.Unmarshal(body, &tmpData)
	if err != nil {
		log.Printf("Failed unmarshalling test: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if tmpData.OrgId != fileId {
		log.Printf("[INFO] OrgId %s and %s don't match", tmpData.OrgId, fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "Organization ID's don't match"}`))
		return
	}

	ctx := GetContext(request)

	org, err := GetOrg(ctx, tmpData.OrgId)
	if err != nil {
		log.Printf("[INFO] Organization %s doesn't exist: %s", tmpData.OrgId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	workflowExecution, err := GetWorkflowExecution(ctx, tmpData.ExecutionRef)
	if err != nil {
		log.Printf("[INFO] Couldn't find workflow execution: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "No permission to get execution"}`))
		return
	}

	if workflowExecution.Authorization != tmpData.Authorization {
		// Get the user?

		log.Printf("[INFO] Execution auth %s and %s don't match", workflowExecution.Authorization, tmpData.Authorization)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "Auth doesn't match"}`))
		return
	}

	if workflowExecution.Status != "EXECUTING" {
		log.Printf("[INFO] Workflow isn't executing and shouldn't be searching", workflowExecution.ExecutionId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "Workflow isn't executing"}`))
		return
	}

	if workflowExecution.ExecutionOrg != org.Id {
		log.Printf("[INFO] Org %s wasn't used to execute %s", org.Id, workflowExecution.ExecutionId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "Bad organization specified"}`))
		return
	}

	// Prepared for the future~
	if len(tmpData.Values) != 1 {
		log.Printf("[INFO] Filter data can only hande 1 value right now, not %d", len(tmpData.Values))
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "Can't handle multiple apps yet, just one"}`))
		return
	}

	value := tmpData.Values[0]

	// FIXME: Alphabetically sort the parameternames
	// FIXME: Add organization wide search, not just workflow based

	found := []string{}
	notFound := []string{}

	dbKey := fmt.Sprintf("app_execution_values")
	parameterNames := fmt.Sprintf("%s_%s", value.App, strings.Join(value.ParameterNames, "_"))
	if tmpData.WorkflowCheck {
		// FIXME: Make this alphabetical
		for _, value := range value.ParameterValues {
			if len(value) == 0 {
				log.Printf("Shouldn't have value of length 0!")
				continue
			}

			log.Printf("[INFO] Looking for value %s in Workflow %s of ORG %s", value, workflowExecution.Workflow.ID, org.Id)

			executionValues, err := GetAppExecutionValues(ctx, parameterNames, org.Id, workflowExecution.Workflow.ID, value)
			if err != nil {
				log.Printf("[WARNING] Failed getting key %s: %s", dbKey, err)
				notFound = append(notFound, value)
				//found = append(found, value)
				continue
			}

			foundCount := len(executionValues)

			if foundCount > 0 {
				found = append(found, value)
			} else {
				log.Printf("[INFO] Found for %s: %d", dbKey, foundCount)
				notFound = append(notFound, value)
			}
		}
	} else {
		log.Printf("[INFO] Should validate if value %s is in workflow id %s", value, workflowExecution.Workflow.ID)
		for _, value := range value.ParameterValues {
			if len(value) == 0 {
				log.Printf("Shouldn't have value of length 0!")
				continue
			}

			log.Printf("[INFO] Looking for value %s in ORG %s", value, org.Id)

			executionValues, err := GetAppExecutionValues(ctx, parameterNames, org.Id, workflowExecution.Workflow.ID, value)
			if err != nil {
				log.Printf("[WARNING] Failed getting key %s: %s", dbKey, err)
				notFound = append(notFound, value)
				//found = append(found, value)
				continue
			}

			foundCount := len(executionValues)

			if foundCount > 0 {
				found = append(found, value)
			} else {
				log.Printf("[INFO] Found for %s: %d", dbKey, foundCount)
				notFound = append(notFound, value)
			}
		}
	}

	//App             string
	//Actions         string
	//ParameterNames  string
	//ParamererValues []string

	appended := 0
	if tmpData.Append {
		log.Printf("[INFO] Should append %d value(s) in K:V for %s_%s!", len(notFound), org.Id, workflowExecution.ExecutionId)

		//parameterNames := strings.Join(value.ParameterNames, "_")
		for _, notFoundValue := range notFound {
			newRequest := NewValue{
				OrgId:               org.Id,
				WorkflowExecutionId: workflowExecution.ExecutionId,
				ParameterName:       parameterNames,
				Value:               notFoundValue,
			}

			// WorkflowId:          workflowExecution.Workflow.Id,
			if tmpData.WorkflowCheck {
				newRequest.WorkflowId = workflowExecution.Workflow.ID
			}

			err = SetNewValue(ctx, newRequest)
			if err != nil {
				log.Printf("[ERROR] Error adding %s to appvalue: %s", notFoundValue, err)
				continue
			}

			appended += 1
			log.Printf("[INFO] Added %s as new appvalue to datastore", notFoundValue)
		}
	}

	type returnStruct struct {
		Success  bool     `json:"success"`
		Appended int      `json:"appended"`
		Found    []string `json:"found"`
	}

	returnData := returnStruct{
		Success:  true,
		Appended: appended,
		Found:    found,
	}

	b, _ := json.Marshal(returnData)
	resp.WriteHeader(200)
	resp.Write(b)
}

// Used for swapping your own organization to a new one IF it's eligible
func HandleChangeUserOrg(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Just getting here for later
	ctx := GetContext(request)
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in change org: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains

		// Clean up the users' cache for different parts

		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {

			DeleteCache(ctx, fmt.Sprintf("%s_workflows", user.Id))
			DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
			DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
			DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))

			log.Printf("[DEBUG] Redirecting ORGCHANGE request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)

			DeleteCache(ctx, fmt.Sprintf("%s_workflows", user.Id))
			DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
			DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
			DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))

			return
		}
	}

	go GetAllWorkflowsByQuery(context.Background(), user)
	go GetPrioritizedApps(context.Background(), user)

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed reading body"}`))
		return
	}

	type ReturnData struct {
		OrgId     string `json:"org_id" datastore:"org_id"`
		RegionUrl string `json:"region_url" datastore:"region_url"`
	}

	var tmpData ReturnData
	err = json.Unmarshal(body, &tmpData)
	if err != nil {
		log.Printf("Failed unmarshalling test: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	foundOrg := false
	for _, org := range user.Orgs {
		if org == tmpData.OrgId {
			foundOrg = true
			break
		}
	}

	if !foundOrg || tmpData.OrgId != fileId {
		log.Printf("[WARNING] User swap to the org \"%s\" - access denied", tmpData.OrgId)
		resp.WriteHeader(403)
		resp.Write([]byte(`{"success": false, "No permission to change to this org"}`))
		return
	}

	org, err := GetOrg(ctx, tmpData.OrgId)
	if err != nil {
		log.Printf("[WARNING] Organization %s doesn't exist: %s", tmpData.OrgId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Add instantswap of backend
	// This could in theory be built out open source as well
	regionUrl := ""
	if project.Environment == "cloud" {
		regionUrl = "https://shuffler.io"
	}

	if project.Environment == "cloud" && len(org.RegionUrl) > 0 && !strings.Contains(org.RegionUrl, "\"") {
		regionUrl = org.RegionUrl
	}

	userFound := false
	usr := User{}
	for _, orgUsr := range org.Users {
		if user.Id == orgUsr.Id {
			usr = orgUsr
			userFound = true
			break
		}
	}

	if !userFound {
		log.Printf("[WARNING] User %s (%s) can't change to org %s (%s) (2)", user.Username, user.Id, org.Name, org.Id)
		resp.WriteHeader(403)
		resp.Write([]byte(`{"success": false, "No permission to change to this org"}`))
		return
	}

	user.ActiveOrg = OrgMini{
		Name: org.Name,
		Id:   org.Id,
		Role: usr.Role,
	}

	user.Role = usr.Role

	err = SetUser(ctx, &user, false)
	if err != nil {
		log.Printf("[ERROR] Failed updating user when changing org: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	expiration := time.Now().Add(3600 * time.Second)

	newCookie := &http.Cookie{
		Name:    "session_token",
		Value:   user.Session,
		Expires: expiration,
	}

	if project.Environment == "cloud" {
		newCookie.Domain = ".shuffler.io"
		newCookie.Secure = true
		newCookie.HttpOnly = true
	}

	http.SetCookie(resp, newCookie)

	newCookie.Name = "__session"
	http.SetCookie(resp, newCookie)

	// Cleanup cache for the user
	DeleteCache(ctx, fmt.Sprintf("%s_workflows", user.Id))
	DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
	DeleteCache(ctx, fmt.Sprintf("apps_%s", user.ActiveOrg.Id))
	DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
	DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))

	log.Printf("[INFO] User %s (%s) successfully changed org to %s (%s)", user.Username, user.Id, org.Name, org.Id)
	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true, "reason": "Successfully added new suborg. Refresh to see it.", "region_url": "%s"}`, regionUrl)))

}

func HandleCreateSubOrg(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in creating sub org: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role != "admin" {
		log.Printf("[WARNING] Can't make suborg without being admin: %s (%s).", user.Username, user.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Not admin"}`))
		return
	}

	ctx := GetContext(request)
	parentOrg, err := GetOrg(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[ERROR] Organization %s doesn't exist or failed to load: %s", user.ActiveOrg.Id, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Just for cache reseting across regions
	for _, inneruser := range parentOrg.Users {
		DeleteCache(ctx, inneruser.ApiKey)
		DeleteCache(ctx, inneruser.Session)
		DeleteCache(ctx, fmt.Sprintf("session_%s", inneruser.Session))
		DeleteCache(ctx, fmt.Sprintf("user_%s", inneruser.Id))
		DeleteCache(ctx, fmt.Sprintf("%s_workflows", inneruser.Id))
		DeleteCache(ctx, fmt.Sprintf("apps_%s", inneruser.Id))
		DeleteCache(ctx, fmt.Sprintf("user_%s", inneruser.Username))
		DeleteCache(ctx, fmt.Sprintf("user_%s", inneruser.Id))
	}

	// Checking if it's a special region. All user-specific requests should
	// go through shuffler.io and not subdomains
	if project.Environment == "cloud" {
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting Create Suborg request to main site handler (shuffler.io)")

			RedirectUserRequest(resp, request)
			return
		}
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed reading body"}`))
		return
	}

	type ReturnData struct {
		OrgId string `json:"org_id" datastore:"org_id"`
		Name  string `json:"name" datastore:"name"`
	}

	var tmpData ReturnData
	err = json.Unmarshal(body, &tmpData)
	if err != nil {
		log.Printf("Failed unmarshalling test: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if len(tmpData.Name) < 3 {
		log.Printf("[WARNING] Suborgname too short (min 3) %s", tmpData.Name)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Name must at least be 3 characters."}`))
		return
	}

	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if tmpData.OrgId != user.ActiveOrg.Id || fileId != user.ActiveOrg.Id {
		log.Printf("[WARNING] User can't edit the org \"%s\"", tmpData.OrgId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "No permission to edit this org (1)"}`))
		return
	}

	if len(parentOrg.ManagerOrgs) > 0 {
		log.Printf("[WARNING] Organization %s can't have suborgs, as it's as suborg", tmpData.OrgId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Can't make suborg of suborg. Switch to a parent org to make another."}`))
		return
	}

	if project.Environment == "cloud" {
		if !parentOrg.SyncFeatures.MultiTenant.Active {
			log.Printf("[WARNING] Org %s is not allowed to make a sub-organization: %s", tmpData.OrgId, err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Sub-organizations require an active subscription with access to multi-tenancy. Contact support to try it out."}`))
			return
		}

		/*
			if parentOrg.SyncUsage.MultiTenant.Counter >= parentOrg.SyncFeatures.MultiTenant.Limit || len(parentOrg.ChildOrgs) > int(parentOrg.SyncFeatures.MultiTenant.Limit) {
				log.Printf("[WARNING] Org %s is not allowed to make ANOTHER sub-organization. Limit reached!: %s", tmpData.OrgId, err)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false, "reason": "Your limit of sub-organizations has been reached. Contact support to increase."}`))
				return
			}
		*/

		parentOrg.SyncUsage.MultiTenant.Counter += 1
		log.Printf("[DEBUG] Allowing suborg for %s because they have %d vs %d limit", parentOrg.Id, len(parentOrg.ChildOrgs), parentOrg.SyncFeatures.MultiTenant.Limit)
	}

	orgId := uuid.NewV4().String()
	newOrg := Org{
		Name:        tmpData.Name,
		Description: fmt.Sprintf("Sub-org by user %s in parent-org %s", user.Username, parentOrg.Name),
		Image:       parentOrg.Image,
		Id:          orgId,
		Org:         tmpData.Name,
		Users: []User{
			user,
		},
		Roles:     []string{"admin", "user"},
		CloudSync: false,
		ManagerOrgs: []OrgMini{
			OrgMini{
				Id:   tmpData.OrgId,
				Name: parentOrg.Name,
			},
		},
		CloudSyncActive: parentOrg.CloudSyncActive,
		CreatorOrg:      tmpData.OrgId,
		ActiveApps:      parentOrg.ActiveApps,
		Region:          parentOrg.Region,
		RegionUrl:       parentOrg.RegionUrl,
	}
	//SyncFeatures:    parentOrg.SyncFeatures,

	parentOrg.ChildOrgs = append(parentOrg.ChildOrgs, OrgMini{
		Name: tmpData.Name,
		Id:   orgId,
	})

	err = SetOrg(ctx, *parentOrg, parentOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Failed updating parent org %s: %s", newOrg.Id, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Update all admins to have access to this suborg
	for _, loopUser := range parentOrg.Users {
		if loopUser.Role != "admin" {
			continue
		}

		if loopUser.Id == user.Id {
			continue
		}

		foundUser, err := GetUser(ctx, loopUser.Id)
		if err != nil {
			log.Printf("[ERROR] User with Identifier %s doesn't exist: %s (update admins - create)", loopUser.Id, err)
			continue
		}

		// Add org to user
		foundUser.Orgs = append(foundUser.Orgs, newOrg.Id)
		err = SetUser(ctx, foundUser, false)
		if err != nil {
			log.Printf("[ERROR] Failed updating user when setting creating suborg (update admins - update): %s ", err)
			continue
		}

		// Add user to org
		newOrg.Users = append(newOrg.Users, loopUser)
	}

	err = SetOrg(ctx, newOrg, newOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Failed setting new org %s: %s", newOrg.Id, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	user.Orgs = append(user.Orgs, newOrg.Id)
	log.Printf("[INFO] Usr: %s (%d)", user.Orgs, len(user.Orgs))
	err = SetUser(ctx, &user, false)
	if err != nil {
		log.Printf("[WARNING] Failed updating user when setting creating suborg: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	log.Printf("[INFO] User %s SUCCESSFULLY ADDED child org %s (%s) for parent %s (%s)", user.Username, newOrg.Name, newOrg.Id, parentOrg.Name, parentOrg.Id)
	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true, "reason": "Successfully updated org"}`)))

}

func HandleEditOrg(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Checking if it's a special region. All user-specific requests should
	// go through shuffler.io and not subdomains
	if project.Environment == "cloud" {
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting Edit Org request to main site handler (shuffler.io)")

			RedirectUserRequest(resp, request)
			return
		}
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in edit org: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role != "admin" {
		log.Printf("[WARNING] Not admin: %s (%s).", user.Username, user.Id)
		resp.WriteHeader(403)
		resp.Write([]byte(`{"success": false, "reason": "Not admin"}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed reading body"}`))
		return
	}

	type ReturnData struct {
		Tutorial    string    `json:"tutorial" datastore:"tutorial"`
		Name        string    `json:"name" datastore:"name"`
		Image       string    `json:"image" datastore:"image"`
		CompanyType string    `json:"company_type" datastore:"company_type"`
		Description string    `json:"description" datastore:"description"`
		OrgId       string    `json:"org_id" datastore:"org_id"`
		Priority    string    `json:"priority" datastore:"priority"`
		Defaults    Defaults  `json:"defaults" datastore:"defaults"`
		SSOConfig   SSOConfig `json:"sso_config" datastore:"sso_config"`
		LeadInfo    []string  `json:"lead_info" datastore:"lead_info"`
	}

	var tmpData ReturnData
	err = json.Unmarshal(body, &tmpData)
	if err != nil {
		log.Printf("Failed unmarshalling test: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}
	//log.Printf("SSO: %s", tmpData.SSOConfig)

	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	admin := false
	if tmpData.OrgId != user.ActiveOrg.Id || fileId != user.ActiveOrg.Id {
		log.Printf("[WARNING] User can't edit org %s (active: %s)", fileId, user.ActiveOrg.Id)
		if !user.SupportAccess {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "No permission to edit this org (2)"}`))
			return
		}

		log.Printf("[AUDIT] User %s (%s) is editing org %s (%s) with support access", user.Username, user.Id, fileId, user.ActiveOrg.Id)
		admin = true
	}

	ctx := GetContext(request)
	org, err := GetOrg(ctx, tmpData.OrgId)
	if err != nil {
		log.Printf("[WARNING] Organization doesn't exist: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	userFound := false
	for _, inneruser := range org.Users {
		if inneruser.Id == user.Id {
			userFound = true
			if inneruser.Role == "admin" {
				admin = true
			}

			break
		}
	}

	if !userFound && !user.SupportAccess {
		log.Printf("[WARNING] User %s doesn't exist in organization for edit %s", user.Id, org.Id)
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if !admin {
		log.Printf("[WARNING] User %s doesn't have edit rights to %s", user.Id, org.Id)
		resp.WriteHeader(403)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	sendOrgUpdaterHook := false
	if len(tmpData.Image) > 0 {
		org.Image = tmpData.Image
	}

	if len(tmpData.Name) > 0 {
		org.Name = tmpData.Name
	}

	if len(tmpData.Description) > 0 {
		org.Description = tmpData.Description
	}

	if len(tmpData.Defaults.AppDownloadRepo) > 0 || len(tmpData.Defaults.AppDownloadBranch) > 0 || len(tmpData.Defaults.WorkflowDownloadRepo) > 0 || len(tmpData.Defaults.WorkflowDownloadBranch) > 0 || len(tmpData.Defaults.NotificationWorkflow) > 0 {
		org.Defaults = tmpData.Defaults
	}

	if len(tmpData.CompanyType) > 0 {
		org.CompanyType = tmpData.CompanyType

		if len(org.CompanyType) == 0 {
			sendOrgUpdaterHook = true
		}
	}

	if len(tmpData.Tutorial) > 0 {
		if tmpData.Tutorial == "welcome" {
			sendOrgUpdaterHook = true
		}
	}

	/*
		// Old code that had frontend buttons.
		// Now we discover this instead
			if len(tmpData.Priority) > 0 {
				if len(org.MainPriority) == 0 {
					org.MainPriority = tmpData.Priority
					sendOrgUpdaterHook = true
				}

				found := false
				for _, prio := range org.Priorities {
					if prio.Name == tmpData.Priority {
						found = true
					}
				}

				if !found {
					org.Priorities = append(org.Priorities, Priority{
						Name:        tmpData.Priority,
						Description: fmt.Sprintf("Priority %s decided by user.", tmpData.Priority),
						Type:        "usecases",
						Active:      true,
						URL:         fmt.Sprintf("/usecases"),
					})
				}
			}
	*/

	//if len(tmpData.SSOConfig) > 0 {
	if len(tmpData.SSOConfig.SSOEntrypoint) > 0 || len(tmpData.SSOConfig.OpenIdClientId) > 0 || len(tmpData.SSOConfig.OpenIdClientSecret) > 0 || len(tmpData.SSOConfig.OpenIdAuthorization) > 0 || len(tmpData.SSOConfig.OpenIdToken) > 0 {
		org.SSOConfig = tmpData.SSOConfig
	}

	if len(tmpData.SSOConfig.SSOCertificate) > 0 {
		savedCert := fixCertificate(tmpData.SSOConfig.SSOCertificate)

		log.Printf("[INFO] Stripped down cert from %d to %d", len(tmpData.SSOConfig.SSOCertificate), len(savedCert))

		org.SSOConfig.SSOCertificate = savedCert
	}

	if len(org.Defaults.NotificationWorkflow) > 0 && len(org.Defaults.NotificationWorkflow) != 36 {
		log.Printf("[WARNING] Notification Workflow ID %s is not valid.", org.Defaults.NotificationWorkflow)
	}

	if len(tmpData.LeadInfo) > 0 && user.SupportAccess {
		log.Printf("[INFO] Updating lead info for %s to %s", org.Id, tmpData.LeadInfo)

		for _, lead := range tmpData.LeadInfo {
			if lead == "contacted" {
				org.LeadInfo.Contacted = true
			}

			if lead == "student" {
				org.LeadInfo.Student = true
			}

			if lead == "lead" {
				org.LeadInfo.Lead = true
			}

			if lead == "pov" {
				org.LeadInfo.POV = true
			}

			if lead == "demo started" {
				org.LeadInfo.DemoDone = true
			}

			if lead == "customer" {
				org.LeadInfo.Customer = true
			}
		}

	}

	// Built a system around this now, which checks for the actual org. Only works onprem so far.
	// if requestdata.Environment == "cloud" && project.Environment != "cloud" {
	//if project.Environment != "cloud" && len(org.SSOConfig.SSOEntrypoint) > 0 && len(org.ManagerOrgs) == 0 {
	//	//log.Printf("[INFO] Should set SSO entrypoint to %s", org.SSOConfig.SSOEntrypoint)
	//	SSOUrl = org.SSOConfig.SSOEntrypoint
	//}

	//log.Printf("Org: %s", org)
	err = SetOrg(ctx, *org, org.Id)
	if err != nil {
		log.Printf("User %s doesn't have edit rights to %s", user.Id, org.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Sends tracker for this on cloud
	if sendOrgUpdaterHook && project.Environment == "cloud" {
		signupWebhook := os.Getenv("WEBSITE_ORG_WEBHOOK")
		if strings.HasPrefix(signupWebhook, "http") {
			curIndex := -1
			for orgIndex, orguser := range org.Users {
				if orguser.Id == user.Id {
					curIndex = orgIndex
				}
			}

			if curIndex >= 0 {
				user.Password = ""
				user.Session = ""
				user.ApiKey = ""
				user.LoginInfo = []LoginInfo{}
				user.PrivateApps = []WorkflowApp{}

				org.Users[curIndex] = user
			}

			mappedData, err := json.Marshal(org)
			if err != nil {
				log.Printf("[WARNING] Marshal error for org sending: %s", err)
			} else {
				req, err := http.NewRequest(
					"POST",
					signupWebhook,
					bytes.NewBuffer(mappedData),
				)

				client := &http.Client{
					Timeout: 3 * time.Second,
				}

				req.Header.Add("Content-Type", "application/json")
				res, err := client.Do(req)
				if err != nil {
					log.Printf("[ERROR] Failed request to signup webhook FOR ORG (2): %s", err)
				} else {
					log.Printf("[INFO] Successfully ran org priority webhook")
				}

				_ = res
			}
		}
	}

	GetTutorials(ctx, *org, true)

	log.Printf("[INFO] Successfully updated org %s (%s) with %d priorities", org.Name, org.Id, len(org.Priorities))
	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true, "reason": "Successfully updated org"}`)))

}

func CheckWorkflowApp(workflowApp WorkflowApp) error {
	// Validate fields
	if workflowApp.Name == "" {
		return errors.New("App field name doesn't exist")
	}

	if workflowApp.Description == "" {
		return errors.New("App field description doesn't exist")
	}

	if workflowApp.AppVersion == "" {
		return errors.New("App field app_version doesn't exist")
	}

	if workflowApp.ContactInfo.Name == "" {
		return errors.New("App field contact_info.name doesn't exist")
	}

	return nil
}

func AbortExecution(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if len(fileId) != 36 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow ID to abort is not valid"}`))
		return
	}

	executionId := location[6]
	if len(executionId) != 36 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "ExecutionID not valid"}`))
		return
	}

	ctx := GetContext(request)
	workflowExecution, err := GetWorkflowExecution(ctx, executionId)
	if err != nil {
		log.Printf("[ERROR] Failed getting execution (abort) %s: %s", executionId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed getting execution ID %s because it doesn't exist (abort)."}`, executionId)))
		return
	}

	apikey := request.Header.Get("Authorization")
	parsedKey := ""
	if strings.HasPrefix(apikey, "Bearer ") {
		apikeyCheck := strings.Split(apikey, " ")
		if len(apikeyCheck) == 2 {
			parsedKey = apikeyCheck[1]
		}
	}

	// Checks the users' role and such if the key fails
	//log.Printf("Abort info: %s vs %s", workflowExecution.Authorization, parsedKey)
	if workflowExecution.Authorization != parsedKey {
		user, err := HandleApiAuthentication(resp, request)
		if err != nil {
			log.Printf("[AUDIT] Api authentication failed in abort workflow: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		//log.Printf("User: %s, org: %s vs %s", user.Role, workflowExecution.Workflow.OrgId, user.ActiveOrg.Id)
		if user.Id != workflowExecution.Workflow.Owner {
			if workflowExecution.Workflow.OrgId == user.ActiveOrg.Id && user.Role == "admin" {
				log.Printf("[AUDIT] User %s is aborting execution %s as admin", user.Username, workflowExecution.Workflow.ID)
			} else {
				log.Printf("[AUDIT] Wrong user (%s) for ABORT of workflowexecution workflow %s", user.Username, workflowExecution.Workflow.ID)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false}`))
				return
			}
		}
	} else {
		//log.Printf("[INFO] API key to abort/finish execution %s is correct.", executionId)
	}

	if workflowExecution.Status == "ABORTED" || workflowExecution.Status == "FAILURE" || workflowExecution.Status == "FINISHED" {
		//err = SetWorkflowExecution(ctx, *workflowExecution, true)
		//if err != nil {
		//}
		log.Printf("[INFO] Stopped execution of %s with status %s", executionId, workflowExecution.Status)
		if len(workflowExecution.ExecutionParent) > 0 {
		}

		//ExecutionSource    string         `json:"execution_source" datastore:"execution_source"`
		//ExecutionParent    string         `json:"execution_parent" datastore:"execution_parent"`

		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Status for %s is %s, which can't be aborted."}`, executionId, workflowExecution.Status)))
		return
	}

	topic := "workflowexecution"

	workflowExecution.CompletedAt = int64(time.Now().Unix())
	workflowExecution.Status = "ABORTED"
	log.Printf("[INFO] Running shutdown (abort) of execution %s", workflowExecution.ExecutionId)

	lastResult := ""
	newResults := []ActionResult{}
	// type ActionResult struct {
	for _, result := range workflowExecution.Results {
		if result.Status == "EXECUTING" {
			result.Status = "ABORTED"
			result.Result = "Aborted because of error in another node (1)"
		}

		if len(result.Result) > 0 && result.Status == "SUCCESS" {
			lastResult = result.Result
		}

		newResults = append(newResults, result)
	}

	workflowExecution.Results = newResults
	if len(workflowExecution.Result) == 0 {
		workflowExecution.Result = lastResult
	}

	addResult := true
	for _, result := range workflowExecution.Results {
		if result.Status != "SKIPPED" {
			addResult = false
		}
	}

	extra := 0
	for _, trigger := range workflowExecution.Workflow.Triggers {
		//log.Printf("Appname trigger (0): %s", trigger.AppName)
		if trigger.AppName == "User Input" || trigger.AppName == "Shuffle Workflow" {
			extra += 1
		}
	}

	parsedReason := "An error occurred during execution of this node. This may be due to the Workflow being Aborted, or an error in the node itself."
	reason, reasonok := request.URL.Query()["reason"]
	if reasonok {
		parsedReason = reason[0]

		// Custom reason handler for weird inputs
		if strings.Contains(parsedReason, "manifest for registry") {
			foundImageSplit := strings.Split(parsedReason, " ")
			foundImage := ""
			if len(foundImageSplit) > 7 {
				foundImage = foundImageSplit[6]

				foundImageSplit = strings.Split(foundImageSplit[6], ":")
				if len(foundImageSplit) > 1 {
					foundImage = foundImageSplit[1]
				}
			}

			parsedReason = fmt.Sprintf("Couldn't find the Docker image %s. Did you activate the app?", foundImage)
		}
	}

	returnData := SubflowData{
		Success: false,
		Result:  parsedReason,
	}

	reasonData, err := json.Marshal(returnData)
	if err != nil {
		reasonData = []byte(parsedReason)
	}

	if len(workflowExecution.Results) == 0 || addResult {
		newaction := Action{
			ID: workflowExecution.Start,
		}

		for _, action := range workflowExecution.Workflow.Actions {
			if action.ID == workflowExecution.Start {
				newaction = action
				break
			}
		}

		workflowExecution.Results = append(workflowExecution.Results, ActionResult{
			Action:        newaction,
			ExecutionId:   workflowExecution.ExecutionId,
			Authorization: workflowExecution.Authorization,
			Result:        string(reasonData),
			StartedAt:     workflowExecution.StartedAt,
			CompletedAt:   workflowExecution.StartedAt,
			Status:        "FAILURE",
		})
	} else if len(workflowExecution.Results) >= len(workflowExecution.Workflow.Actions)+extra {
		log.Printf("[INFO] DONE - Nothing to add during abort!")
	} else {
		//log.Printf("VALIDATING INPUT!")
		node, nodeok := request.URL.Query()["node"]
		if nodeok {
			nodeId := node[0]
			log.Printf("[INFO] Found abort node %s", nodeId)
			newaction := Action{
				ID: nodeId,
			}

			// Check if result exists first
			found := false
			for _, result := range workflowExecution.Results {
				if result.Action.ID == nodeId {
					found = true
					break
				}
			}

			if !found {
				for _, action := range workflowExecution.Workflow.Actions {
					if action.ID == nodeId {
						newaction = action
						break
					}
				}

				workflowExecution.Results = append(workflowExecution.Results, ActionResult{
					Action:        newaction,
					ExecutionId:   workflowExecution.ExecutionId,
					Authorization: workflowExecution.Authorization,
					Result:        string(reasonData),
					StartedAt:     workflowExecution.StartedAt,
					CompletedAt:   workflowExecution.StartedAt,
					Status:        "FAILURE",
				})
			}
		}
	}

	for resultIndex, result := range workflowExecution.Results {
		for parameterIndex, param := range result.Action.Parameters {
			if param.Configuration {
				workflowExecution.Results[resultIndex].Action.Parameters[parameterIndex].Value = ""
			}
		}
	}

	for actionIndex, action := range workflowExecution.Workflow.Actions {
		for parameterIndex, param := range action.Parameters {
			if param.Configuration {
				//log.Printf("Cleaning up %s in %s", param.Name, action.Name)
				workflowExecution.Workflow.Actions[actionIndex].Parameters[parameterIndex].Value = ""
			}
		}
	}

	// This is the same as aborted
	IncrementCache(ctx, workflowExecution.ExecutionOrg, "workflow_executions_failed")
	err = SetWorkflowExecution(ctx, *workflowExecution, true)
	if err != nil {
		log.Printf("[WARNING] Error saving workflow execution for updates when aborting (2) %s: %s", topic, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed setting workflowexecution status to abort"}`)))
		return
	} else {
		log.Printf("[INFO] Set workflowexecution %s to aborted.", workflowExecution.ExecutionId)
	}

	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true}`)))
}

func SanitizeWorkflow(workflow Workflow) Workflow {
	log.Printf("[INFO] Sanitizing workflow %s", workflow.ID)

	for _, trigger := range workflow.Triggers {
		_ = trigger
	}

	for _, action := range workflow.Actions {
		_ = action
	}

	for _, variable := range workflow.WorkflowVariables {
		_ = variable
	}

	workflow.Owner = ""
	workflow.Org = []OrgMini{}
	workflow.OrgId = ""
	workflow.ExecutingOrg = OrgMini{}
	workflow.PreviouslySaved = false

	// Add Gitguardian or similar secrets discovery
	return workflow
}

// Starts a new webhook
func HandleNewHook(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in set new hook: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to make new hook: %s (%s)", user.Username, user.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	type requestData struct {
		Type           string `json:"type"`
		Description    string `json:"description"`
		Id             string `json:"id"`
		Name           string `json:"name"`
		Workflow       string `json:"workflow"`
		Start          string `json:"start"`
		Environment    string `json:"environment"`
		Auth           string `json:"auth"`
		CustomResponse string `json:"custom_response"`
		Version        string `json:"version" datastore:"version"`
		VersionTimeout int    `json:"version_timeout" datastore:"version_timeout"`
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Body data error in webhook set: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)
	var requestdata requestData
	err = json.Unmarshal([]byte(body), &requestdata)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshaling inputdata for webhook: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	newId := requestdata.Id
	if len(newId) != 36 {
		log.Printf("[WARNING] Bad webhook ID: %s", newId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Invalid Webhook ID: bad formatting"}`))
		return
	}

	if requestdata.Id == "" || requestdata.Name == "" {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Required fields id and name can't be empty"}`))
		return

	}

	validTypes := []string{
		"webhook",
	}

	isTypeValid := false
	for _, thistype := range validTypes {
		if requestdata.Type == thistype {
			isTypeValid = true
			break
		}
	}

	if !(isTypeValid) {
		log.Printf("[WARNING] Type %s is not valid. Try any of these: %s", requestdata.Type, strings.Join(validTypes, ", "))
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Let remote endpoint handle access checks (shuffler.io)
	baseUrl := "https://shuffler.io"
	if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
		baseUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
	}

	if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
		baseUrl = os.Getenv("SHUFFLE_CLOUDRUN_URL")
	}

	currentUrl := fmt.Sprintf("%s/api/v1/hooks/webhook_%s", baseUrl, newId)
	startNode := requestdata.Start
	if requestdata.Environment == "cloud" && project.Environment != "cloud" {
		// https://shuffler.io/v1/hooks/webhook_80184973-3e82-4852-842e-0290f7f34d7c
		log.Printf("[INFO] Should START a cloud webhook for url %s for startnode %s", currentUrl, startNode)
		org, err := GetOrg(ctx, user.ActiveOrg.Id)
		if err != nil {
			log.Printf("Failed finding org %s: %s", org.Id, err)
			return
		}

		action := CloudSyncJob{
			Type:          "webhook",
			Action:        "start",
			OrgId:         org.Id,
			PrimaryItemId: newId,
			SecondaryItem: startNode,
			ThirdItem:     requestdata.Workflow,
			FourthItem:    requestdata.Auth,
		}

		err = executeCloudAction(action, org.SyncConfig.Apikey)
		if err != nil {
			log.Printf("[WARNING] Failed cloud action START webhook execution: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
			return
		} else {
			log.Printf("[INFO] Successfully set up cloud action schedule")
		}
	}

	hook := Hook{
		Id:        newId,
		Start:     startNode,
		Workflows: []string{requestdata.Workflow},
		Info: Info{
			Name:        requestdata.Name,
			Description: requestdata.Description,
			Url:         fmt.Sprintf("%s/api/v1/hooks/webhook_%s", baseUrl, newId),
		},
		Type:   "webhook",
		Owner:  user.Username,
		Status: "uninitialized",
		Actions: []HookAction{
			HookAction{
				Type:  "workflow",
				Name:  requestdata.Name,
				Id:    requestdata.Workflow,
				Field: "",
			},
		},
		Running:        false,
		OrgId:          user.ActiveOrg.Id,
		Environment:    requestdata.Environment,
		Auth:           requestdata.Auth,
		CustomResponse: requestdata.CustomResponse,
		Version:        requestdata.Version,
		VersionTimeout: requestdata.VersionTimeout,
	}

	hook.Status = "running"
	hook.Running = true
	err = SetHook(ctx, hook)
	if err != nil {
		log.Printf("[WARNING] Failed setting hook: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	log.Printf("[INFO] Set up a new hook with ID %s and environment %s", newId, hook.Environment)
	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

func HandleDeleteHook(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in delete hook: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to delete hook: %s (%s)", user.Username, user.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")

	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if len(fileId) != 36 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow ID when deleting hook is not valid"}`))
		return
	}

	ctx := GetContext(request)
	hook, err := GetHook(ctx, fileId)
	if err != nil {
		log.Printf("[WARNING] Failed getting hook %s (delete): %s", fileId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	//if user.Id != hook.Owner && user.ActiveOrg.Id != hook.OrgId {
	//	log.Printf("[WARNING] Wrong user (%s) for workflow %s", user.Username, hook.Id)
	//	resp.WriteHeader(401)
	//	resp.Write([]byte(`{"success": false}`))
	//	return
	//}

	if user.Id != hook.Owner || len(user.Id) == 0 {
		if hook.OrgId == user.ActiveOrg.Id && user.Role == "admin" {
			log.Printf("[AUDIT] User %s is stopping hook for workflow %s as admin. Owner: %s", user.Username, hook.Workflows[0], hook.Owner)
		} else {
			log.Printf("[AUDIT] Wrong user (%s) for hook %s (stop hook)", user.Username, hook.Workflows[0])
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	if len(hook.Workflows) > 0 {
		//err = increaseStatisticsField(ctx, "total_workflow_triggers", hook.Workflows[0], -1, user.ActiveOrg.Id)
		//if err != nil {
		//	log.Printf("Failed to increase total workflows: %s", err)
		//}
	}

	hook.Status = "stopped"
	err = SetHook(ctx, *hook)
	if err != nil {
		log.Printf("[WARNING] Failed setting hook: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if hook.Environment == "cloud" && project.Environment != "cloud" {
		log.Printf("[INFO] Should STOP cloud webhook https://shuffler.io/api/v1/hooks/webhook_%s", hook.Id)
		org, err := GetOrg(ctx, user.ActiveOrg.Id)
		if err != nil {
			log.Printf("Failed finding org %s: %s", org.Id, err)
			return
		}

		action := CloudSyncJob{
			Type:          "webhook",
			Action:        "stop",
			OrgId:         org.Id,
			PrimaryItemId: hook.Id,
		}

		if len(hook.Workflows) > 0 {
			action.SecondaryItem = hook.Workflows[0]
		}

		err = executeCloudAction(action, org.SyncConfig.Apikey)
		if err != nil {
			log.Printf("Failed cloud action STOP execution: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
			return
		}
		// https://shuffler.io/v1/hooks/webhook_80184973-3e82-4852-842e-0290f7f34d7c
	}

	err = DeleteKey(ctx, "hooks", fileId)
	if err != nil {
		log.Printf("[WARNING] Error deleting hook %s for %s: %s", fileId, user.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed deleting the hook."}`))
		return
	}

	log.Printf("[INFO] Successfully deleted webhook %s", fileId)
	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true, "reason": "Stopped webhook"}`))
}

func ParseVersions(versions []string) []string {
	log.Printf("Versions: %s", versions)

	//versions = sort.Sort(semver.Collection(versions))
	return versions
}

func GetWorkflowAppConfig(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting App request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	log.Printf("[INFO] Running GetWorkflowAppConfig for '%s'", fileId)

	ctx := GetContext(request)
	app, err := GetApp(ctx, fileId, User{}, false)
	if err != nil {
		log.Printf("[WARNING] Error getting app %s (app config): %s", fileId, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "App doesn't exist"}`))
		return
	}

	log.Printf("[INFO] Successfully got app %s", fileId)

	app.ReferenceUrl = ""

	type AppParser struct {
		Success bool   `json:"success"`
		OpenAPI []byte `json:"openapi"`
		App     []byte `json:"app"`
	}

	//app.Activate = true
	data, err := json.Marshal(app)
	if err != nil {
		resp.WriteHeader(422)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed marshalling new parsed APP: %s"}`, err)))
		return
	}

	appReturn := AppParser{
		Success: true,
		App:     data,
	}

	appdata, err := json.Marshal(appReturn)
	if err != nil {
		log.Printf("[WARNING] Error parsing appReturn for app (INIT): %s", err)
		resp.WriteHeader(422)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed marshalling: %s"}`, err)))
		return
	}

	user, userErr := HandleApiAuthentication(resp, request)
	log.Printf("[INFO] User: %s", user.Username)

	openapi, openapiok := request.URL.Query()["openapi"]
	//if app.Sharing || app.Public || (project.Environment == "cloud" && user.Id == "what") {
	//log.Printf("SHARING: %s. PUBLIC: %s", app.Sharing, app.Public)
	if app.Sharing || app.Public {
		if openapiok && len(openapi) > 0 && strings.ToLower(openapi[0]) == "false" {
			log.Printf("Should return WITHOUT openapi")
		} else {
			//log.Printf("CAN SHARE APP!")
			parsedApi, err := GetOpenApiDatastore(ctx, fileId)
			if err != nil {
				log.Printf("[WARNING] OpenApi doesn't exist for (0): %s - err: %s. Returning basic app", fileId, err)
				resp.WriteHeader(200)
				resp.Write(appdata)
				return
			}

			if len(parsedApi.Body) > 0 {
				if len(parsedApi.ID) > 0 {
					parsedApi.Success = true
				} else {
					parsedApi.Success = false
				}

				//log.Printf("PARSEDAPI: %s", parsedApi)
				openapidata, err := json.Marshal(parsedApi)
				if err != nil {
					log.Printf("[WARNING] Error parsing api json: %s", err)
					resp.WriteHeader(422)
					resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed marshalling new parsed swagger: %s"}`, err)))
					return
				}

				appReturn.OpenAPI = openapidata
			}
		}

		appdata, err = json.Marshal(appReturn)
		if err != nil {
			log.Printf("[WARNING] Error parsing appReturn for app: %s", err)
			resp.WriteHeader(422)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed marshalling: %s"}`, err)))
			return
		}

		resp.WriteHeader(200)
		resp.Write(appdata)
		return
	}

	if userErr != nil {
		log.Printf("[WARNING] Api authentication failed in get app: %s", userErr)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Modified to make it so users admins in same org can modify an app
	//log.Printf("User: %s, role: %s, org: %s vs %s", user.Username, user.Role, user.ActiveOrg.Id, app.ReferenceOrg)
	if user.Id == app.Owner || (user.Role == "admin" && user.ActiveOrg.Id == app.ReferenceOrg) {
		log.Printf("[DEBUG] Got app %s with user %s (%s) in org %s", app.ID, user.Username, user.Id, user.ActiveOrg.Id)
	} else {
		if project.Environment == "cloud" && user.Verified == true && user.Active == true && user.SupportAccess == true && strings.HasSuffix(user.Username, "@shuffler.io") {
			log.Printf("[AUDIT] Support & Admin user %s (%s) got access to app %s (cloud only)", user.Username, user.Id, app.ID)
		} else if user.Role == "admin" && app.Owner == "" {
			log.Printf("[AUDIT] Any admin can GET %s (%s), since it doesn't have an owner (GET).", app.Name, app.ID)
		} else {
			exit := true

			log.Printf("[INFO] Check published: %s", app.PublishedId)
			if len(app.PublishedId) > 0 {

				// FIXME: is this privacy / vulnerability?
				// Allows parent owner to see child usage.
				// Intended to allow vision of changes, and have parent app suggestions be possible
				parentapp, err := GetApp(ctx, app.PublishedId, user, false)
				if err == nil {
					if parentapp.Owner == user.Id {
						log.Printf("[AUDIT] Parent app owner %s got access to child app %s (%s)", user.Username, user.Id, app.Name, app.ID)
						exit = false
						//app, err := GetApp(ctx, fileId, User{}, false)
					}
				}
			}

			if exit {
				log.Printf("[AUDIT] Wrong user (%s) for app %s (%s) - get app config", user.Username, app.Name, app.ID)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false}`))
				return
			}
		}
	}

	if openapiok && len(openapi) > 0 && strings.ToLower(openapi[0]) == "false" {
		//log.Printf("Should return WITHOUT openapi")
	} else {
		log.Printf("[INFO] Getting app %s (OpenAPI)", fileId)
		parsedApi, err := GetOpenApiDatastore(ctx, fileId)
		if err != nil {
			log.Printf("[INFO] OpenApi doesn't exist for (1): %s - err: %s. Returning basic app.", fileId, err)

			resp.WriteHeader(200)
			resp.Write(appdata)
			return
		}

		if len(parsedApi.ID) > 0 {
			parsedApi.Success = true
		} else {
			parsedApi.Success = false
		}

		openapidata, err := json.Marshal(parsedApi)
		if err != nil {
			resp.WriteHeader(422)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed marshalling new parsed swagger: %s"}`, err)))
			return
		}

		appReturn.OpenAPI = openapidata
	}

	appdata, err = json.Marshal(appReturn)
	if err != nil {
		log.Printf("[WARNING] Error parsing appReturn for app: %s", err)
		resp.WriteHeader(422)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed marshalling: %s"}`, err)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(appdata)
}

func verifier() (*CodeVerifier, error) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 32, 32)
	for i := 0; i < 32; i++ {
		b[i] = byte(r.Intn(255))
	}
	return CreateCodeVerifierFromBytes(b)
}

func GetOpenIdUrl(request *http.Request, org Org) string {
	baseSSOUrl := org.SSOConfig.OpenIdAuthorization

	codeChallenge := uuid.NewV4().String()
	//h.Write([]byte(v.Value))
	verifier, verifiererr := verifier()
	if verifiererr == nil {
		codeChallenge = verifier.Value
	}

	//log.Printf("[DEBUG] Got challenge value %s (pre state)", codeChallenge)

	// https://192.168.55.222:3443/api/v1/login_openid
	//location := strings.Split(request.URL.String(), "/")
	//redirectUrl := url.QueryEscape("http://localhost:5001/api/v1/login_openid")
	redirectUrl := url.QueryEscape(fmt.Sprintf("http://%s/api/v1/login_openid", request.Host))
	if project.Environment == "cloud" {
		redirectUrl = url.QueryEscape(fmt.Sprintf("https://shuffler.io/api/v1/login_openid"))
	}

	if project.Environment != "cloud" && strings.Contains(request.Host, "shuffle-backend") && !strings.Contains(os.Getenv("BASE_URL"), "shuffle-backend") {
		redirectUrl = url.QueryEscape(fmt.Sprintf("%s/api/v1/login_openid", os.Getenv("BASE_URL")))
	}

	if project.Environment != "cloud" && len(os.Getenv("SSO_REDIRECT_URL")) > 0 {
		redirectUrl = url.QueryEscape(fmt.Sprintf("%s/api/v1/login_openid", os.Getenv("SSO_REDIRECT_URL")))
	}

	state := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("org=%s&challenge=%s&redirect=%s", org.Id, codeChallenge, redirectUrl)))

	// has to happen after initial value is stored
	if verifiererr == nil {
		codeChallenge = verifier.CodeChallengeS256()
	}

	//log.Printf("[DEBUG] Got challenge value %s (POST state)", codeChallenge)

	if len(org.SSOConfig.OpenIdClientSecret) > 0 {

		//baseSSOUrl += fmt.Sprintf("?client_id=%s&response_type=code&scope=openid&redirect_uri=%s&state=%s&client_secret=%s", org.SSOConfig.OpenIdClientId, redirectUrl, state, org.SSOConfig.OpenIdClientSecret)
		state := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("org=%s&redirect=%s&challenge=%s", org.Id, redirectUrl, org.SSOConfig.OpenIdClientSecret)))
		log.Printf("[INFO] URL: %s", redirectUrl)

		baseSSOUrl += fmt.Sprintf("?client_id=%s&response_type=id_token&scope=openid&redirect_uri=%s&state=%s&response_mode=form_post&nonce=%s", org.SSOConfig.OpenIdClientId, redirectUrl, state, state)
		//baseSSOUrl += fmt.Sprintf("&client_secret=%s", org.SSOConfig.OpenIdClientSecret)
		log.Printf("[DEBUG] Found OpenID url (client secret). Extra redirect check: %s - %s", request.URL.String(), baseSSOUrl)
	} else {
		log.Printf("[DEBUG] Found OpenID url (PKCE!!). Extra redirect check: %s", request.URL.String())
		baseSSOUrl += fmt.Sprintf("?client_id=%s&response_type=code&scope=openid&redirect_uri=%s&state=%s&code_challenge_method=S256&code_challenge=%s", org.SSOConfig.OpenIdClientId, redirectUrl, state, codeChallenge)
	}

	return baseSSOUrl
}

func HandleLogin(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	if project.Environment == "cloud" {
		// Checking if it's a special region. All user-specific requests should
		// go through shuffler.io and not subdomains
		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		if gceProject != "shuffler" && gceProject != sandboxProject && len(gceProject) > 0 {
			log.Printf("[DEBUG] Redirecting LOGIN request to main site handler (shuffler.io)")
			RedirectUserRequest(resp, request)
			return
		}
	}

	// Gets a struct of Username, password
	data, err := ParseLoginParameters(resp, request)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
		return
	}

	log.Printf("[AUDIT] Handling login of %s", data.Username)

	data.Username = strings.ToLower(strings.TrimSpace(data.Username))
	err = checkUsername(data.Username)
	if err != nil {
		log.Printf("[INFO] Username is too short or bad for %s: %s", data.Username, err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
		return
	}

	ctx := GetContext(request)
	users, err := FindUser(ctx, data.Username)
	if err != nil && len(users) == 0 {
		log.Printf("[WARNING] Failed getting user %s during login", data.Username)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Username and/or password is incorrect"}`))
		return
	}

	userdata := User{}
	if len(users) != 1 {
		log.Printf("[WARNING] Username %s has multiple or no users (%d). Checking if it matches any.", data.Username, len(users))

		for _, user := range users {
			if user.Id == "" && user.Username == "" {
				log.Printf(`[AUDIT] Username %s (%s) isn't valid (2). Amount of users checked: %d (1)`, user.Username, user.Id, len(users))
				continue
			}

			if user.ActiveOrg.Id != "" {
				err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
				if err != nil {
					log.Printf("[WARNING] Bad password: %s", err)
					continue
				}

				userdata = user
				break
			}
		}
	} else {
		userdata = users[0]
	}

	// Starting caching of the username
	// This is to make it faster later :)
	go GetAllWorkflowsByQuery(context.Background(), userdata)
	go GetPrioritizedApps(context.Background(), userdata)

	/*
			// FIXME: Reenable activation?
		if project.Environment == "cloud" && !userdata.Active {
			log.Printf("[DEBUG] %s is not active, but tried to login. Error: %v", data.Username, err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "This user is deactivated"}`))
			return
		}
	*/

	org := Org{}
	updateUser := false
	if project.Environment == "cloud" {
		if strings.HasSuffix(strings.ToLower(userdata.Username), "@shuffler.io") {
			if !userdata.Active {
				log.Printf("[INFO] User %s with @shuffler suffix is not active.", userdata.Username)
				resp.WriteHeader(401)
				resp.Write([]byte(fmt.Sprintf(`{"success": true, "reason": "error: You need to activate your account before logging in"}`)))
				return
			}
		}

		//log.Printf("[DEBUG] Are they using SSO?")
		// If it fails, allow login if password correct?
		// Check if suborg -> Get parent & check SSO
		baseOrg, err := GetOrg(ctx, userdata.ActiveOrg.Id)
		if err == nil {
			//log.Printf("Got org during signin: %s - checking SAML SSO", baseOrg.Id)
			org = *baseOrg
			if len(baseOrg.ManagerOrgs) > 0 {

				// Use auth from parent org if user is also in that one
				newOrg, err := GetOrg(ctx, baseOrg.ManagerOrgs[0].Id)
				if err == nil {

					found := false
					for _, user := range newOrg.Users {
						if user.Username == userdata.Username {
							found = true
						}
					}

					if found {
						log.Printf("[WARNING] Using parent org of %s as org %s", baseOrg.Id, newOrg.Id)
						org = *newOrg
					}
				}
			}

			log.Printf("[INFO] Inside SSO / OpenID check: %s", org.Id)
			if len(org.SSOConfig.SSOEntrypoint) > 0 || len(org.SSOConfig.OpenIdAuthorization) > 0 {
				baseSSOUrl := org.SSOConfig.SSOEntrypoint
				redirectKey := "SSO_REDIRECT"
				if len(org.SSOConfig.OpenIdAuthorization) > 0 {
					log.Printf("[INFO] OpenID login for %s", org.Id)
					redirectKey = "SSO_REDIRECT"

					baseSSOUrl = GetOpenIdUrl(request, org)
				}

				log.Printf("[DEBUG] Should redirect user %s in org %s(%s) to SSO login at %s", userdata.Username, userdata.ActiveOrg.Name, userdata.ActiveOrg.Id, baseSSOUrl)

				// Check if the user has other orgs that can be swapped to - if so SWAP
				userDomain := strings.Split(userdata.Username, "@")
				for _, tmporg := range userdata.Orgs {
					innerorg, err := GetOrg(ctx, tmporg)
					if err != nil {
						continue
					}

					if innerorg.Id == userdata.ActiveOrg.Id {
						continue
					}

					if len(innerorg.ManagerOrgs) > 0 {
						continue
					}

					// Not your own org
					if innerorg.Org == userdata.Username || strings.Contains(innerorg.Name, "@") {
						continue
					}

					if len(userDomain) >= 2 {
						if strings.Contains(strings.ToLower(innerorg.Org), strings.ToLower(userDomain[1])) {
							continue
						}
					}

					// Shouldn't contain the domain of the users' email
					log.Printf("[ERROR] Found org for %s (%s) to check into instead of running OpenID/SSO: %s.", userdata.Username, userdata.Id, innerorg.Name)
					userdata.ActiveOrg.Id = innerorg.Id
					userdata.ActiveOrg.Name = innerorg.Name

					DeleteCache(ctx, fmt.Sprintf("%s_workflows", userdata.Id))
					DeleteCache(ctx, fmt.Sprintf("apps_%s", userdata.Id))
					DeleteCache(ctx, fmt.Sprintf("user_%s", userdata.Username))
					DeleteCache(ctx, fmt.Sprintf("user_%s", userdata.Id))

					updateUser = true
					break

				}

				// user controllable field hmm :)
				if !updateUser {
					resp.WriteHeader(401)
					resp.Write([]byte(fmt.Sprintf(`{"success": true, "reason": "%s", "url": "%s"}`, redirectKey, baseSSOUrl)))
					return
				}
			}
		}
	}

	if len(users) == 1 {
		err = bcrypt.CompareHashAndPassword([]byte(userdata.Password), []byte(data.Password))
		if err != nil {
			userdata = User{}
			log.Printf("[WARNING] Bad password: %s", err)
		} else {
			log.Printf("[DEBUG] Correct password with single user!")
		}
	}

	if userdata.Id == "" && userdata.Username == "" {
		log.Printf(`[AUDIT] Login for Username %s isn't valid with that password. Amount of users checked: %d (2)`, data.Username, len(users))
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Username and/or password is incorrect"}`)))
		return
	}

	if updateUser {
		err = SetUser(ctx, &userdata, false)
		if err != nil {
			log.Printf("[WARNING] Failed updating user when auto-setting new org: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Something went wrong with the SSO redirect system"}`)))
			return
		}
	}

	if userdata.LoginType == "SSO" {
		log.Printf(`[WARNING] Username %s (%s) has login type set to SSO (single sign-on).`, userdata.Username, userdata.Id)
		//resp.WriteHeader(401)
		//resp.Write([]byte(`{"success": false, "reason": "This user can only log in with SSO"}`))
		//return
	}

	if userdata.LoginType == "OpenID" {
		log.Printf(`[WARNING] Username %s (%s) has login type set to OpenID (single sign-on).`, userdata.Username, userdata.Id)
	}

	if userdata.MFA.Active && len(data.MFACode) == 0 {
		log.Printf(`[DEBUG] Username %s (%s) has MFA activated. Redirecting.`, userdata.Username, userdata.Id)
		resp.WriteHeader(409)
		resp.Write([]byte(fmt.Sprintf(`{"success": true, "reason": "MFA_REDIRECT"}`)))
		return
	}

	if len(data.MFACode) > 0 && userdata.MFA.Active {
		interval := time.Now().Unix() / 30
		HOTP, err := getHOTPToken(userdata.MFA.ActiveCode, interval)
		if err != nil {
			log.Printf("[ERROR] Failed generating a HOTP token: %s", err)
			resp.WriteHeader(500)
			resp.Write([]byte(`{"success": false, "reason": "Failed generating token. Please try again."}`))
			return
		}

		if HOTP != data.MFACode {
			log.Printf("[DEBUG] Bad code sent for user %s (%s). Sent: %s, Want: %s", userdata.Username, userdata.Id, data.MFACode, HOTP)
			resp.WriteHeader(500)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Wrong 2-factor code (%s). Please try again with a 6-digit code. If this persists, please contact support."}`, data.MFACode)))
			return
		}

		log.Printf("[DEBUG] MFA login for user %s (%s)!", userdata.Username, userdata.Id)
	}

	//tutorialsFinished := userdata.PersonalInfo.Tutorials
	//if len(org.SecurityFramework.SIEM.Name) > 0 || len(org.SecurityFramework.Network.Name) > 0 || len(org.SecurityFramework.EDR.Name) > 0 || len(org.SecurityFramework.Cases.Name) > 0 || len(org.SecurityFramework.IAM.Name) > 0 || len(org.SecurityFramework.Assets.Name) > 0 || len(org.SecurityFramework.Intel.Name) > 0 || len(org.SecurityFramework.Communication.Name) > 0 {
	//	tutorialsFinished = append(tutorialsFinished, "find_integrations")
	//}

	userdata.LoginInfo = append(userdata.LoginInfo, LoginInfo{
		IP:        request.RemoteAddr,
		Timestamp: time.Now().Unix(),
	})

	tutorialsFinished := []Tutorial{}
	for _, tutorial := range userdata.PersonalInfo.Tutorials {
		tutorialsFinished = append(tutorialsFinished, Tutorial{
			Name: tutorial,
		})
	}

	if len(org.Id) == 0 {
		newOrg, err := GetOrg(ctx, userdata.ActiveOrg.Id)
		if err == nil {
			org = *newOrg
		}
	}

	if len(org.SecurityFramework.SIEM.Name) > 0 || len(org.SecurityFramework.Network.Name) > 0 || len(org.SecurityFramework.EDR.Name) > 0 || len(org.SecurityFramework.Cases.Name) > 0 || len(org.SecurityFramework.IAM.Name) > 0 || len(org.SecurityFramework.Assets.Name) > 0 || len(org.SecurityFramework.Intel.Name) > 0 || len(org.SecurityFramework.Communication.Name) > 0 {
		tutorialsFinished = append(tutorialsFinished, Tutorial{
			Name: "find_integrations",
		})
	}

	for _, tutorial := range org.Tutorials {
		tutorialsFinished = append(tutorialsFinished, tutorial)
	}

	//log.Printf("[INFO] Tutorials finished: %v", tutorialsFinished)

	returnValue := HandleInfo{
		Success:   true,
		Tutorials: tutorialsFinished,
	}

	loginData := `{"success": true}`
	newData, err := json.Marshal(returnValue)
	if err == nil {
		loginData = string(newData)
	}

	if len(userdata.Session) != 0 {
		log.Printf("[INFO] User session exists - resetting session")
		expiration := time.Now().Add(3600 * time.Second)

		newCookie := &http.Cookie{
			Name:    "session_token",
			Value:   userdata.Session,
			Expires: expiration,
		}

		if project.Environment == "cloud" {
			newCookie.Domain = ".shuffler.io"
			newCookie.Secure = true
			newCookie.HttpOnly = true
		}

		http.SetCookie(resp, newCookie)

		newCookie.Name = "__session"
		http.SetCookie(resp, newCookie)

		//log.Printf("SESSION LENGTH MORE THAN 0 IN LOGIN: %s", userdata.Session)
		returnValue.Cookies = append(returnValue.Cookies, SessionCookie{
			Key:        "session_token",
			Value:      userdata.Session,
			Expiration: expiration.Unix(),
		})

		returnValue.Cookies = append(returnValue.Cookies, SessionCookie{
			Key:        "__session",
			Value:      userdata.Session,
			Expiration: expiration.Unix(),
		})

		loginData = fmt.Sprintf(`{"success": true, "cookies": [{"key": "session_token", "value": "%s", "expiration": %d}]}`, userdata.Session, expiration.Unix())
		newData, err := json.Marshal(returnValue)
		if err == nil {
			loginData = string(newData)
		}

		err = SetSession(ctx, userdata, userdata.Session)
		if err != nil {
			log.Printf("[WARNING] Error adding session to database: %s", err)
		} else {
			//log.Printf("[DEBUG] Updated session in backend")
		}

		err = SetUser(ctx, &userdata, false)
		if err != nil {
			log.Printf("[ERROR] Failed updating user when setting session (2): %s", err)
			resp.WriteHeader(500)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		resp.WriteHeader(200)
		resp.Write([]byte(loginData))
		return
	} else {
		log.Printf("[INFO] User session for %s (%s) is empty - create one!", userdata.Username, userdata.Id)

		sessionToken := uuid.NewV4().String()
		expiration := time.Now().Add(3600 * time.Second)
		newCookie := &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: expiration,
		}

		if project.Environment == "cloud" {
			newCookie.Domain = ".shuffler.io"
			newCookie.Secure = true
			newCookie.HttpOnly = true
		}

		// Does it not set both?
		http.SetCookie(resp, newCookie)

		newCookie.Name = "__session"
		http.SetCookie(resp, newCookie)

		// ADD TO DATABASE
		err = SetSession(ctx, userdata, sessionToken)
		if err != nil {
			log.Printf("[DEBUG] Error adding session to database: %s", err)
		}

		userdata.Session = sessionToken

		returnValue.Cookies = append(returnValue.Cookies, SessionCookie{
			Key:        "session_token",
			Value:      sessionToken,
			Expiration: expiration.Unix(),
		})

		returnValue.Cookies = append(returnValue.Cookies, SessionCookie{
			Key:        "__session",
			Value:      sessionToken,
			Expiration: expiration.Unix(),
		})

		err = SetUser(ctx, &userdata, true)
		if err != nil {
			log.Printf("[ERROR] Failed updating user when setting session: %s", err)
			resp.WriteHeader(500)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		loginData = fmt.Sprintf(`{"success": true, "cookies": [{"key": "session_token", "value": "%s", "expiration": %d}]}`, sessionToken, expiration.Unix())
		newData, err := json.Marshal(returnValue)
		if err == nil {
			loginData = string(newData)
		}
	}

	log.Printf("[INFO] %s SUCCESSFULLY LOGGED IN with session %s", data.Username, userdata.Session)

	resp.WriteHeader(200)
	resp.Write([]byte(loginData))
}

func ParseLoginParameters(resp http.ResponseWriter, request *http.Request) (loginStruct, error) {
	if request.Body == nil {
		return loginStruct{}, errors.New("Failed to parse login params, body is empty")
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return loginStruct{}, err
	}

	var t loginStruct

	err = json.Unmarshal(body, &t)
	if err != nil {
		return loginStruct{}, err
	}

	return t, nil
}

func checkUsername(Username string) error {
	// Stupid first check of email loool
	//if !strings.Contains(Username, "@") || !strings.Contains(Username, ".") {
	//	return errors.New("Invalid Username")
	//}

	if len(Username) < 3 {
		return errors.New("Minimum Username length is 3")
	}

	return nil
}

// Handles workflow executions across systems (open source, worker, cloud)
// getWorkflow
// GetWorkflow
// executeWorkflow

// This should happen locally.. Meaning, polling may be stupid.
// Let's do it anyway, since it seems like the best way to scale
// without remoting problems and the like.
func updateExecutionParent(ctx context.Context, executionParent, returnValue, parentAuth, parentNode, subflowExecutionId string) error {

	// Was an error here. Now defined to run with http://shuffle-backend:5001 by default
	backendUrl := os.Getenv("BASE_URL")
	if project.Environment == "cloud" {
		backendUrl = "https://shuffler.io"

		if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
			backendUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
		}

		if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
			backendUrl = os.Getenv("SHUFFLE_CLOUDRUN_URL")
		}

		//backendUrl = "http://localhost:5002"
	}

	// FIXME: This MAY fail at scale due to not being able to get the right worker
	// Maybe we need to pass the worker's real id, and not its VIP?
	if os.Getenv("SHUFFLE_SWARM_CONFIG") == "run" && (project.Environment == "" || project.Environment == "worker") {
		backendUrl = "http://shuffle-workers:33333"

		hostenv := os.Getenv("WORKER_HOSTNAME")
		if len(hostenv) > 0 {
			backendUrl = fmt.Sprintf("http://%s:33333", hostenv)
		}

		// From worker:
		//parsedRequest.BaseUrl = fmt.Sprintf("http://%s:%d", hostname, baseport)

		log.Printf("[DEBUG][%s] Sending request for shuffle-subflow result to %s. Should this be a specific worker? Specific worker is better if cache is NOT memcached", subflowExecutionId, backendUrl)
	}

	//log.Printf("[INFO] PARENTEXEC: %s, AUTH: %s, parentNode: %s, BackendURL: %s, VALUE: %s. ", executionParent, parentAuth, parentNode, backendUrl, returnValue)

	// Callback to itself
	if len(backendUrl) == 0 {
		backendUrl = "http://localhost:5001"
	}

	resultUrl := fmt.Sprintf("%s/api/v1/streams/results", backendUrl)
	//log.Printf("[DEBUG] ResultURL: %s", backendUrl)
	topClient := &http.Client{
		Transport: &http.Transport{
			Proxy: nil,
		},
	}
	newExecution := WorkflowExecution{}

	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")
	if len(httpProxy) > 0 || len(httpsProxy) > 0 {
		topClient = &http.Client{}
	} else {
		if len(httpProxy) > 0 {
			log.Printf("[INFO] Running with HTTP proxy %s (env: HTTP_PROXY)", httpProxy)
		}
		if len(httpsProxy) > 0 {
			log.Printf("[INFO] Running with HTTPS proxy %s (env: HTTPS_PROXY)", httpsProxy)
		}
	}

	requestData := ActionResult{
		Authorization: parentAuth,
		ExecutionId:   executionParent,
	}

	data, err := json.Marshal(requestData)
	if err != nil {
		log.Printf("[WARNING] Failed parent init marshal: %s", err)
		return err
	}

	req, err := http.NewRequest(
		"POST",
		resultUrl,
		bytes.NewBuffer([]byte(data)),
	)

	newresp, err := topClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed making parent request: %s. Is URL valid: %s", err, backendUrl)
		return err
	}

	body, err := ioutil.ReadAll(newresp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed reading parent body: %s", err)
		return err
	}
	//log.Printf("BODY (%d): %s", newresp.StatusCode, string(body))

	if newresp.StatusCode != 200 {
		log.Printf("[ERROR] Bad statuscode setting subresult (1) with URL %s: %d, %s. Input data: %s", resultUrl, newresp.StatusCode, string(body), string(data))
		return errors.New(fmt.Sprintf("Bad statuscode: %s", newresp.StatusCode))
	}

	err = json.Unmarshal(body, &newExecution)
	if err != nil {
		log.Printf("[ERROR] Failed newexecutuion parent unmarshal: %s", err)
		return err
	}

	foundResult := ActionResult{}
	for _, result := range newExecution.Results {
		if result.Action.ID == parentNode {
			foundResult = result
			break
		}
	}

	//log.Printf("FOUND RESULT: %s", foundResult)
	isLooping := false
	selectedTrigger := Trigger{}
	for _, trigger := range newExecution.Workflow.Triggers {
		if trigger.ID == parentNode {
			selectedTrigger = trigger
			for _, param := range trigger.Parameters {
				if param.Name == "argument" && strings.Contains(param.Value, "$") && strings.Contains(param.Value, ".#") {
					isLooping = true
					break
				}
			}

			break
		}
	}

	// Because we changed out how we handle mid-flow triggers
	if len(selectedTrigger.ID) == 0 {
		for _, action := range newExecution.Workflow.Actions {
			if action.ID == parentNode {
				selectedTrigger = Trigger{
					ID:    action.ID,
					Label: action.Label,
				}

				foundResult.Action = action

				for _, param := range action.Parameters {
					if param.Name == "argument" && strings.Contains(param.Value, "$") && strings.Contains(param.Value, ".#") {
						isLooping = true
						break
					}
				}

				break
			}
		}
	}

	// IF the workflow is looping, the result is added in the backend to not
	// cause consistency issues. This means the result will be sent back, and instead
	// Added to the workflow result by the backend itself.
	// When all the "WAITING" executions are done, the backend will set the execution itself
	// back to executing, allowing the parent to continue
	sendRequest := false
	resultData := []byte{}
	if isLooping {
		//log.Printf("\n\n[DEBUG] ITS LOOPING - SHOULD ADD TO A LIST INSTEAD!\n\n")

		// Saved for each subflow ID -> parentNode
		subflowResultCacheId := fmt.Sprintf("%s_%s_subflowresult", subflowExecutionId, parentNode)
		err = SetCache(ctx, subflowResultCacheId, []byte(returnValue), 30)
		if err != nil {
			log.Printf("\n\n\n[ERROR] Failed setting subflow loop cache result for action in parsed exec results %s: %s\n\n", subflowResultCacheId, err)
			return err
		}

		// Every time we get here, we need to both SET the value in cache AND look for other values in cache to make sure the list is good.
		parentNodeFound := false
		var parentSubflowResult []SubflowData
		for _, result := range newExecution.Results {
			if result.Action.ID == parentNode {
				//log.Printf("[DEBUG] FOUND RES: %s", foundResult.Result)

				parentNodeFound = true
				err = json.Unmarshal([]byte(foundResult.Result), &parentSubflowResult)
				if err != nil {
					log.Printf("[ERROR] Failed to unmarshal result to parentsubflow res: %s", err)
					continue
				}

				break
			}
		}

		// If found, loop through and make sure to check the result for ALL of them. If they're not in there, add them as values.
		if parentNodeFound {
			log.Printf("[DEBUG] Found result for subflow (parentNodeFound). Got %d parentSubflowResults", len(parentSubflowResult))

			ranUpdate := false

			newResults := []SubflowData{}
			finishedSubflows := 0
			for _, res := range parentSubflowResult {
				// If value length = 0 for any, then check cache and add the result
				//log.Printf("[DEBUG] EXEC: %s", res)
				if res.ExecutionId == subflowExecutionId {
					//foundResult.Result
					res.Result = string(returnValue)
					res.ResultSet = true

					ranUpdate = true

					//log.Printf("[DEBUG] Set the result for the node! Run update with %s", res)
					finishedSubflows += 1
				} else {
					res.ResultSet = true

					// Overutilization of cache :>
					if !res.ResultSet || len(res.Result) == 0 {
						subflowResultCacheId = fmt.Sprintf("%s_%s_subflowresult", res.ExecutionId, parentNode)

						cache, err := GetCache(ctx, subflowResultCacheId)
						if err == nil {
							cacheData := []byte(cache.([]uint8))
							//log.Printf("[DEBUG] Cachedata for other subflow: %s", string(cacheData))
							res.Result = string(cacheData)
							res.ResultSet = true
							ranUpdate = true

							finishedSubflows += 1
						} else {
							//log.Printf("[DEBUG] No cache data set for subflow cache %s", subflowResultCacheId)
						}
					} else {
						finishedSubflows += 1
					}
				}

				newResults = append(newResults, res)
			}

			if finishedSubflows == len(newResults) {
				log.Printf("[DEBUG] Finished workflow because status of all should be set to finished now")

				// Status is used to determine if the current subflow is finished
				//foundResult.Status = "SUCCESS"
				foundResult.Status = "FINISHED"

			}

			if ranUpdate {

				sendRequest = true
				baseResultData, err := json.Marshal(newResults)
				if err != nil {
					log.Printf("[ERROR] Failed marshalling subflow loop request data (1): %s", err)
					return err
				}

				foundResult.Result = string(baseResultData)
				foundResult.ExecutionId = executionParent
				foundResult.Authorization = parentAuth
				resultData, err = json.Marshal(foundResult)
				if err != nil {
					log.Printf("[ERROR] Failed marshalling FULL subflow loop request data (2): %s", err)
					return err
				}

				//log.Printf("[DEBUG] Should update with multiple results for the subflow. Fullres: %s!", string(foundResult.Result))

			}
		} else {
			log.Printf("[DEBUG] Did not enter parentNodeFound in subflow loop")

		}

		// Check if the item alreayd exists or not in results
		//return nil
	} else {
		//log.Printf("\n\n[DEBUG] ITS NOT LOOP for parent node '%s'. Found data: %s\n\n", parentNode, returnValue)

		if len(selectedTrigger.ID) > 0 {
			foundResult.Action.ID = selectedTrigger.ID
		}

		// 1. Get result of parentnode's subflow (foundResult.Result)
		// 2. Try to marshal parent into a loop.
		// 3. If possible, loop through and find the one matching SubflowData.ExecutionId with "executionParent"
		// 4. If it's matching, update ONLY that one.

		var subflowDataLoop []SubflowData
		err = json.Unmarshal([]byte(foundResult.Result), &subflowDataLoop)
		if err == nil {
			for subflowIndex, subflowData := range subflowDataLoop {
				if subflowData.ExecutionId == executionParent {
					log.Printf("[DEBUG] Updating execution Id %s with subflow info", subflowData.ExecutionId)
					subflowDataLoop[subflowIndex].Result = returnValue
				}
			}

			//foundResult.ExecutionId = executionParent
			//foundResult.Authorization = parentAuth
			resultData, err = json.Marshal(subflowDataLoop)
			if err != nil {
				log.Printf("[WARNING] Failed updating resultData (4): %s", err)
				return err
			}

			sendRequest = true
		} else {
			// In here maning no-loop?
			actionValue := SubflowData{
				Success:       true,
				ExecutionId:   executionParent,
				Authorization: parentAuth,
				Result:        returnValue,
			}

			parsedActionValue, err := json.Marshal(actionValue)
			if err != nil {
				log.Printf("[ERROR] Failed updating resultData (1): %s", err)
				return err
			}

			// This is probably bad for loops
			timeNow := time.Now().Unix()
			if len(foundResult.Action.ID) == 0 {
				log.Printf("\n\n[INFO] Couldn't find the result? Data: %s\n\n", string(resultData))
				parsedAction := Action{
					Label:       selectedTrigger.Label,
					ID:          parentNode,
					Name:        "run_subflow",
					AppName:     "shuffle-subflow",
					AppVersion:  "1.0.0",
					Environment: selectedTrigger.Environment,
				}

				newResult := ActionResult{
					Action:        parsedAction,
					ExecutionId:   executionParent,
					Authorization: parentAuth,
					Result:        string(parsedActionValue),
					StartedAt:     timeNow,
					CompletedAt:   timeNow,
					Status:        "SUCCESS",
				}

				resultData, err = json.Marshal(newResult)
				if err != nil {
					log.Printf("[ERROR] Failed updating resultData (2): %s", err)
					return err
				}

				sendRequest = true
			} else {
				log.Printf("[DEBUG] Found result. Sending result with input data")

				foundResult.StartedAt = timeNow
				foundResult.CompletedAt = timeNow
				foundResult.Authorization = parentAuth
				foundResult.ExecutionId = executionParent
				foundResult.Result = string(parsedActionValue)

				if foundResult.Status == "" {
					foundResult.Status = "SUCCESS"
				}

				resultData, err = json.Marshal(foundResult)
				if err != nil {
					log.Printf("[ERROR] Failed updating resultData (3): %s", err)
					return err
				}

				sendRequest = true
			}
		}
	}

	// This is to ensure cache is in time. Timing issues between parent & child nodes are awful :)
	if isLooping {
		cacheId := fmt.Sprintf("%s_%s_result", foundResult.ExecutionId, foundResult.Action.ID)
		err = SetCache(ctx, cacheId, resultData, 35)
		if err != nil {
			log.Printf("[WARNING] Couldn't set cache for subflow action result %s (4): %s", cacheId, err)
		}

		log.Printf("[DEBUG] Set cache for subflow action result loop %s (4) with 250 ms delay before request", cacheId)
		//time.Sleep(250 * time.Millisecond)
	}

	if sendRequest && len(resultData) > 0 {
		//log.Printf("[INFO][%s] Should send subflow request to backendURL %s. Data: %s!", executionParent, backendUrl, string(resultData))

		streamUrl := fmt.Sprintf("%s/api/v1/streams", backendUrl)
		req, err := http.NewRequest(
			"POST",
			streamUrl,
			bytes.NewBuffer([]byte(resultData)),
		)

		if err != nil {
			log.Printf("Error building subflow request: %s", err)
			return err
		}

		newresp, err := topClient.Do(req)
		if err != nil {
			log.Printf("Error running subflow request: %s", err)
			return err
		}

		defer newresp.Body.Close()
		if newresp.StatusCode != 200 {
			body, err := ioutil.ReadAll(newresp.Body)
			if err != nil {
				log.Printf("[INFO] Failed reading body after subflow request: %s", err)
				return err
			} else {
				log.Printf("[ERROR] Failed forwarding subflow request of length %d\n: %s", len(resultData), string(body))
				//log.Printf("[ERROR] Failed forwarding subflow request %s\n: %s", string(resultData), string(body))
			}
		}
	} else {
		log.Printf("[INFO] NOT sending request because data len is %d and request is %s", len(resultData), sendRequest)
	}

	return nil
}

func ResendActionResult(actionData []byte, retries int64) {
	if project.Environment == "cloud" && retries == 0 {
		retries = 4
		//return

		//var res ActionResult
		//err := json.Unmarshal(actionData, &res)
		//if err == nil {
		//	log.Printf("[WARNING] Cloud - skipping rerun with %d retries for %s (%s)", retries, res.Action.Label, res.Action.ID)
		//}

		//return
	}

	if retries >= 5 {
		return
	}

	backendUrl := os.Getenv("BASE_URL")
	if project.Environment == "cloud" {
		backendUrl = "https://shuffler.io"

		if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
			backendUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
		}

		if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
			backendUrl = os.Getenv("SHUFFLE_CLOUDRUN_URL")
		}
	}

	if os.Getenv("SHUFFLE_SWARM_CONFIG") == "run" && (project.Environment == "" || project.Environment == "worker") {
		backendUrl = "http://shuffle-workers:33333"

		// Should connect to self, not shuffle-workers
		hostenv := os.Getenv("WORKER_HOSTNAME")
		if len(hostenv) > 0 {
			backendUrl = fmt.Sprintf("http://%s:33333", hostenv)
		}
		//parsedRequest.BaseUrl = fmt.Sprintf("http://%s:%d", hostname, baseport)

		// From worker:
		//parsedRequest.BaseUrl = fmt.Sprintf("http://%s:%d", hostname, baseport)

		log.Printf("\n\n[DEBUG] REsending request to rerun action result to %s\n\n", backendUrl)

		// Here to prevent infinite loops
		var res ActionResult
		err := json.Unmarshal(actionData, &res)
		if err == nil {
			ctx := context.Background()
			parsedValue, err := GetBackendexecution(ctx, res.ExecutionId, res.Authorization)
			if err != nil {
				log.Printf("[WARNING] Failed getting execution from backend to verify (3): %s", err)
			} else {
				log.Printf("[INFO][%s] Found execution result (3) %s for subflow %s in backend with %d results and result %s", res.ExecutionId, parsedValue.Status, res.ExecutionId, len(parsedValue.Results), parsedValue.Result)
				if parsedValue.Status != "EXECUTING" {
					return
				}
			}
		}
	}

	if len(backendUrl) == 0 {
		backendUrl = "http://localhost:5001"
	}

	log.Printf("\n\n[INFO] Resending action result to backend %s\n\n", backendUrl)

	streamUrl := fmt.Sprintf("%s/api/v1/streams?rerun=true&retries=%d", backendUrl, retries+1)
	req, err := http.NewRequest(
		"POST",
		streamUrl,
		bytes.NewBuffer(actionData),
	)

	if err != nil {
		log.Printf("[ERROR] Error building resend action request - retries: %d, err: %s", retries, err)

		if project.Environment != "cloud" && retries < 5 {
			if strings.Contains(fmt.Sprintf("%s", err), "cannot assign requested address") {
				time.Sleep(5 * time.Second)
				retries = retries + 1

				ResendActionResult(actionData, retries)
			}
		}

		return
	}

	//Timeout: 3 * time.Second,
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: nil,
		},
	}

	_, err = client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Error running resend action request - retries: %d, err: %s", retries, err)

		if !strings.Contains(fmt.Sprintf("%s", err), "context deadline") && !strings.Contains(fmt.Sprintf("%s", err), "Client.Timeout exceeded") {
			// How to self repair? Quit and restart the worker?
			// This means worker is buggy when talking to itself
			if project.Environment != "cloud" && retries < 5 {
				if strings.Contains(fmt.Sprintf("%s", err), "cannot assign requested address") {
					time.Sleep(5 * time.Second)
					retries = retries + 1

					ResendActionResult(actionData, retries)
				}
			} else if project.Environment != "cloud" && retries >= 5 {
				//panic("No more sockets available. Restarting worker to self-repair.")
				log.Printf("[WARNING] Should we quit out on worker and start a new? How can we remove socket boundry?")
			}
		}

		return
	}

	//body, err := ioutil.ReadAll(newresp.Body)
	//if err != nil {
	//	log.Printf("[WARNING] Error getting body from rerun: %s", err)
	//	return
	//}

	//log.Printf("[DEBUG] Status %d and Body from rerun: %s", newresp.StatusCode, string(body))
}

// Function for translating action results into whatever.
// Came about because of general issues with Oauth2
func FixActionResultOutput(actionResult ActionResult) ActionResult {
	if strings.Contains(actionResult.Result, "TypeError") && strings.Contains(actionResult.Result, "missing 1 required positional argument: 'access_token'") {
		//log.Printf("\n\nTypeError  in actionresult!")
		actionResult.Result = `{"success": false, "reason": "This App requires authentication with Oauth2. Make sure to authenticate it first."}`
	}

	return actionResult
}

// Updateparam is a check to see if the execution should be continuously validated
func ParsedExecutionResult(ctx context.Context, workflowExecution WorkflowExecution, actionResult ActionResult, updateParam bool, retries int64) (*WorkflowExecution, bool, error) {
	var err error
	if actionResult.Action.ID == "" {
		// Can we find it based on label?

		log.Printf("\n\n[ERROR][%s] Failed handling EMPTY action %#v (ParsedExecutionResult). Usually ONLY happens during worker run that sets everything?\n\n", workflowExecution.ExecutionId, actionResult)

		return &workflowExecution, true, nil
	}

	// 1. CHECK cache if it happened in another?
	// 2. Set cache
	// 3. Find executed without a result
	// 4. Ensure the result is NOT set when running an action

	actionResult = FixActionResultOutput(actionResult)
	actionCacheId := fmt.Sprintf("%s_%s_result", actionResult.ExecutionId, actionResult.Action.ID)
	// Done elsewhere

	// Don't set cache for triggers?
	//log.Printf("\n\nACTIONRES: %s\n\nRES: %s\n", actionResult, actionResult.Result)

	setCache := true
	if actionResult.Action.AppName == "shuffle-subflow" {

		// Verifying if the userinput should be sent properly or not
		if actionResult.Action.Name == "run_userinput" {
			//log.Printf("\n\n[INFO] Inside userinput default return! Return data: %s", actionResult.Result)

			if strings.Contains(actionResult.Result, "\"success\":") {
				//log.Printf("Found success in result. Now verifying if the workflow should just continue or not")

				type SubflowMapping struct {
					Success bool `json:"success"`
				}

				var subflowData SubflowMapping
				err := json.Unmarshal([]byte(actionResult.Result), &subflowData)
				if err == nil && subflowData.Success == false {
					log.Printf("[INFO][%s] Userinput subflow failed. Should abort workflow or continue execution by default?", actionResult.ExecutionId)
				} else {
					log.Printf("[INFO][%s] Userinput subflow succeeded. Should continue execution by default? Value: %s", actionResult.ExecutionId, actionResult.Result)
				}
			}

			// Finding the waiting node and changing it to this result
			foundWaiting := false
			for resultIndex, result := range workflowExecution.Results {
				if result.Action.ID == actionResult.Action.ID {
					log.Printf("\n\n[INFO] Found user input to replace.\n\n")
					workflowExecution.Results[resultIndex].Result = actionResult.Result

					// Updating cache for the result to always use the latest
					//actionResultBody, err := json.Marshal(workflowExecution.Results[resultIndex].Result)
					actionResultBody, err := json.Marshal(actionResult)
					if err == nil {
						cacheId := fmt.Sprintf("%s_%s_result", workflowExecution.ExecutionId, actionResult.Action.ID)
						err = SetCache(ctx, cacheId, actionResultBody, 35)
						if err != nil {
							log.Printf("[WARNING] Couldn't find in fix exec %s (2): %s", cacheId, err)
							continue
						}
					}

					foundWaiting = true
					break
				}
			}

			if !foundWaiting {
				workflowExecution.Results = append(workflowExecution.Results, actionResult)

				actionResultBody, err := json.Marshal(actionResult)
				if err == nil {
					cacheId := fmt.Sprintf("%s_%s_result", workflowExecution.ExecutionId, actionResult.Action.ID)
					err = SetCache(ctx, cacheId, actionResultBody, 35)
					if err != nil {
						log.Printf("[ERROR][%s] Failed to update cache for %s", workflowExecution.ExecutionId, cacheId)
					}
				}
			}

			err = SetWorkflowExecution(ctx, workflowExecution, true)
			if err != nil {
				log.Printf("[ERROR][%s] Failed setting workflow execution during user input return: %s", workflowExecution.ExecutionId, err)
			}

			return &workflowExecution, true, nil
		} else {
			for _, param := range actionResult.Action.Parameters {
				if param.Name == "check_result" {
					//log.Printf("[INFO] RESULT: %s", param)
					if param.Value == "true" {
						setCache = false
					}

					break
				}
			}

			if !setCache {
				var subflowData SubflowData
				jsonerr := json.Unmarshal([]byte(actionResult.Result), &subflowData)
				if jsonerr == nil && len(subflowData.Result) == 0 && !strings.Contains(actionResult.Result, "\"result\"") {
					setCache = false
				} else {
					setCache = true
				}
			}
		}

		//log.Printf("[DEBUG] Skipping setcache for subflow? SetCache: %t", setCache)
	}

	if setCache {
		actionResultBody, err := json.Marshal(actionResult)
		if err == nil {
			err = SetCache(ctx, actionCacheId, actionResultBody, 35)
			if err != nil {
				log.Printf("\n\n\n[ERROR] Failed setting cache for action in parsed exec results %s: %s\n\n", actionCacheId, err)
			}
		} else {
			log.Printf("\n\n[ERROR] Failed marshalling result and put it in cache.")
		}
	} else {
		//log.Printf("[WARNING] Skipping cache for %s", actionResult.Action.Name)
	}

	skipExecutionCount := false
	if workflowExecution.Status == "FINISHED" {
		skipExecutionCount = true
	}

	dbSave := false
	startAction, extra, children, parents, visited, executed, nextActions, environments := GetExecutionVariables(ctx, workflowExecution.ExecutionId)

	//log.Printf("RESULT: %s", actionResult.Action.ExecutionVariable)
	// Shitty workaround as it may be missing it at times
	for _, action := range workflowExecution.Workflow.Actions {
		if action.ID == actionResult.Action.ID {
			//log.Printf("HAS EXEC VARIABLE: %s", action.ExecutionVariable)
			actionResult.Action.ExecutionVariable = action.ExecutionVariable
			break
		}
	}

	newResult := FixBadJsonBody([]byte(actionResult.Result))
	actionResult.Result = string(newResult)

	if len(actionResult.Action.ExecutionVariable.Name) > 0 && (actionResult.Status == "SUCCESS" || actionResult.Status == "FINISHED") {

		setExecVar := true
		//log.Printf("\n\n[DEBUG] SETTING ExecVar RESULTS: %s", actionResult.Result)
		if strings.Contains(actionResult.Result, "\"success\":") {
			type SubflowMapping struct {
				Success bool `json:"success"`
			}

			var subflowData SubflowMapping
			err := json.Unmarshal([]byte(actionResult.Result), &subflowData)
			if err != nil {
				log.Printf("[ERROR] Failed to map in set execvar name with success: %s", err)
				setExecVar = false
			} else {
				if subflowData.Success == false {
					setExecVar = false
				}
			}
		}

		if len(actionResult.Result) == 0 {
			setExecVar = false
		}

		if setExecVar {
			log.Printf("[DEBUG][%s] Updating exec variable %s with new value of length %d (2)", workflowExecution.ExecutionId, actionResult.Action.ExecutionVariable.Name, len(actionResult.Result))

			if len(workflowExecution.Results) > 0 {
				// Should this be used?
				// lastResult := workflowExecution.Results[len(workflowExecution.Results)-1].Result
			}

			actionResult.Action.ExecutionVariable.Value = actionResult.Result

			foundIndex := -1
			for i, executionVariable := range workflowExecution.ExecutionVariables {
				if executionVariable.Name == actionResult.Action.ExecutionVariable.Name {
					foundIndex = i
					break
				}
			}

			if foundIndex >= 0 {
				workflowExecution.ExecutionVariables[foundIndex] = actionResult.Action.ExecutionVariable
			} else {
				workflowExecution.ExecutionVariables = append(workflowExecution.ExecutionVariables, actionResult.Action.ExecutionVariable)
			}
		} else {
			log.Printf("[DEBUG] NOT updating exec variable %s with new value of length %d. Checkp revious errors, or if action was successful (success: true)", actionResult.Action.ExecutionVariable.Name, len(actionResult.Result))
		}
	}

	actionResult.Action = Action{
		AppName:           actionResult.Action.AppName,
		AppVersion:        actionResult.Action.AppVersion,
		Label:             actionResult.Action.Label,
		Name:              actionResult.Action.Name,
		ID:                actionResult.Action.ID,
		Parameters:        actionResult.Action.Parameters,
		ExecutionVariable: actionResult.Action.ExecutionVariable,
	}

	// Cleaning up result authentication
	for paramIndex, param := range actionResult.Action.Parameters {
		if param.Configuration {
			//log.Printf("[INFO] Deleting param %s (auth)", param.Name)
			actionResult.Action.Parameters[paramIndex].Value = ""
		}
	}

	// Used for testing subflow shit
	//if strings.Contains(actionResult.Action.Label, "Shuffle Workflow_30") {
	//	log.Printf("RESULT FOR %s: %s", actionResult.Action.Label, actionResult.Result)
	//	if !strings.Contains(actionResult.Result, "\"result\"") {
	//		log.Printf("NO RESULT - RETURNING!")
	//		return &workflowExecution, false, nil
	//	}
	//}

	// Fills in data from subflows, whether they're loops or not
	// Deprecated! Now runs updateExecutionParent() instead
	// Update: handling this farther down the function
	//log.Printf("[DEBUG] STATUS OF %s: %s", actionResult.Action.AppName, actionResult.Status)
	if actionResult.Status == "SUCCESS" && actionResult.Action.AppName == "shuffle-subflow" {
		dbSave = true
	}

	if actionResult.Status == "ABORTED" || actionResult.Status == "FAILURE" {
		IncrementCache(ctx, workflowExecution.ExecutionOrg, "app_executions_failed")

		if workflowExecution.Workflow.Configuration.SkipNotifications == false {
			// Add an else for HTTP request errors with success "false"
			// These could be "silent" issues
			if actionResult.Status == "FAILURE" && workflowExecution.Workflow.Hidden == false {
				log.Printf("[DEBUG] Result is %s for %s (%s). Making notification.", actionResult.Status, actionResult.Action.Label, actionResult.Action.ID)
				err := CreateOrgNotification(
					ctx,
					fmt.Sprintf("Error in Workflow %s", workflowExecution.Workflow.Name),
					fmt.Sprintf("Node %s in Workflow %s was found to have an error. Click to investigate", actionResult.Action.Label, workflowExecution.Workflow.Name),
					fmt.Sprintf("/workflows/%s?execution_id=%s&view=executions&node=%s", workflowExecution.Workflow.ID, workflowExecution.ExecutionId, actionResult.Action.ID),
					workflowExecution.ExecutionOrg,
					true,
				)

				if err != nil {
					log.Printf("[WARNING] Failed making org notification: %s", err)
				}
			}
		}

		newResults := []ActionResult{}
		childNodes := []string{}
		if workflowExecution.Workflow.Configuration.ExitOnError {
			// Find underlying nodes and add them
			log.Printf("[WARNING] Actionresult is %s for node %s (%s) in execution %s. Should set workflowExecution and exit all running functions", actionResult.Status, actionResult.Action.Label, actionResult.Action.ID, workflowExecution.ExecutionId)
			workflowExecution.Status = actionResult.Status
			workflowExecution.LastNode = actionResult.Action.ID

			if len(workflowExecution.Workflow.DefaultReturnValue) > 0 {
				workflowExecution.Result = workflowExecution.Workflow.DefaultReturnValue
			}

			IncrementCache(ctx, workflowExecution.ExecutionOrg, "workflow_executions_failed")
		} else {

			log.Printf("[WARNING] Actionresult is %s for node %s in %s. Continuing anyway because of workflow configuration.", actionResult.Status, actionResult.Action.ID, workflowExecution.ExecutionId)
			// Finds ALL childnodes to set them to SKIPPED
			// Remove duplicates
			//log.Printf("CHILD NODES: %d", len(childNodes))
			childNodes = FindChildNodes(workflowExecution, actionResult.Action.ID, []string{}, []string{})
			//log.Printf("\n\nFOUND %d CHILDNODES\n\n", len(childNodes))
			for _, nodeId := range childNodes {
				if nodeId == actionResult.Action.ID {
					continue
				}

				// 1. Find the action itself
				// 2. Create an actionresult
				curAction := Action{ID: ""}
				for _, action := range workflowExecution.Workflow.Actions {
					if action.ID == nodeId {
						curAction = action
						break
					}
				}

				isTrigger := false
				if len(curAction.ID) == 0 {
					for _, trigger := range workflowExecution.Workflow.Triggers {
						//log.Printf("%s : %s", trigger.ID, nodeId)
						if trigger.ID == nodeId {
							isTrigger = true
							name := "shuffle-subflow"
							curAction = Action{
								AppName:    name,
								AppVersion: trigger.AppVersion,
								Label:      trigger.Label,
								Name:       trigger.Name,
								ID:         trigger.ID,
							}

							//log.Printf("SET NODE!!")
							break
						}
					}

					if len(curAction.ID) == 0 {
						//log.Printf("Couldn't find subnode %s", nodeId)
						continue
					}
				}

				resultExists := false
				for _, result := range workflowExecution.Results {
					if result.Action.ID == curAction.ID {
						resultExists = true
						break
					}
				}

				if !resultExists {
					// Check parents are done here. Only add it IF all parents are skipped
					skipNodeAdd := false
					for _, branch := range workflowExecution.Workflow.Branches {
						if branch.DestinationID == nodeId && !isTrigger {
							// If the branch's source node is NOT in childNodes, it's not a skipped parent
							// Checking if parent is a trigger
							parentTrigger := false
							for _, trigger := range workflowExecution.Workflow.Triggers {
								if trigger.ID == branch.SourceID {
									if trigger.AppName != "User Input" && trigger.AppName != "Shuffle Workflow" {
										parentTrigger = true
									}
								}
							}

							if parentTrigger {
								continue
							}

							sourceNodeFound := false
							for _, item := range childNodes {
								if item == branch.SourceID {
									sourceNodeFound = true
									break
								}
							}

							if !sourceNodeFound {
								// FIXME: Shouldn't add skip for child nodes of these nodes. Check if this node is parent of upcoming nodes.
								//log.Printf("\n\n NOT setting node %s to SKIPPED", nodeId)
								skipNodeAdd = true

								if !ArrayContains(visited, nodeId) && !ArrayContains(executed, nodeId) {
									nextActions = append(nextActions, nodeId)
									log.Printf("[INFO] SHOULD EXECUTE NODE %s. Next actions: %s", nodeId, nextActions)
								}
								break
							}
						}
					}

					if !skipNodeAdd {
						newResult := ActionResult{
							Action:        curAction,
							ExecutionId:   actionResult.ExecutionId,
							Authorization: actionResult.Authorization,
							Result:        `{"success": false, "reason": "Skipped because of previous node - 2"}`,
							StartedAt:     0,
							CompletedAt:   0,
							Status:        "SKIPPED",
						}

						newResults = append(newResults, newResult)

						visited = append(visited, curAction.ID)
						executed = append(executed, curAction.ID)

						UpdateExecutionVariables(ctx, workflowExecution.ExecutionId, startAction, children, parents, visited, executed, nextActions, environments, extra)
					} else {
						//log.Printf("\n\nNOT adding %s as skipaction - should add to execute?", nodeId)
						//var visited []string
						//var executed []string
						//var nextActions []string
					}
				}
			}
		}

		// Cleans up aborted, and always gives a result
		lastResult := ""
		// type ActionResult struct {
		for _, result := range workflowExecution.Results {
			if actionResult.Action.ID == result.Action.ID {
				continue
			}

			if result.Status == "EXECUTING" {
				result.Status = actionResult.Status
				result.Result = "Aborted because of error in another node (2)"
			}

			if len(result.Result) > 0 && result.Status == "SUCCESS" {
				lastResult = result.Result
			}

			newResults = append(newResults, result)
		}

		if workflowExecution.LastNode == "" {
			workflowExecution.LastNode = actionResult.Action.ID
		}

		workflowExecution.Result = lastResult
		workflowExecution.Results = newResults
	}

	if actionResult.Status == "SKIPPED" {
		//childNodes := FindChildNodes(workflowExecution, actionResult.Action.ID)

		// See if it can even find it in here for skipped?
		//log.Printf("childnodes of %s (%s): %d: %s", actionResult.Action.Label, actionResult.Action.Id, len(childNodes), childNodes)

		//FIXME: Should this run and fix all nodes,
		// or should it send them in as new SKIPs? Should we only handle DIRECT
		// children? I wonder.

		//log.Printf("\n\n\n[DEBUG] FROM %s - FOUND childnode %s %s (%s). exists: %s\n\n\n", actionResult.Action.Label, curAction.ID, curAction.Name, curAction.Label, resultExists)
		// FIXME: Add triggers
		for _, branch := range workflowExecution.Workflow.Branches {
			if branch.SourceID != actionResult.Action.ID {
				continue
			}

			// Find the target & check if it has more branches. If it does, and they're not finished - continue
			foundAction := Action{}
			for _, action := range workflowExecution.Workflow.Actions {
				if action.ID == branch.DestinationID {
					foundAction = action
					break
				}
			}

			if len(foundAction.ID) == 0 {
				for _, trigger := range workflowExecution.Workflow.Triggers {
					//if trigger.AppName == "User Input" || trigger.AppName == "Shuffle Workflow" {
					if trigger.ID == branch.DestinationID {
						foundAction = Action{
							ID:      trigger.ID,
							AppName: trigger.AppName,
							Name:    trigger.AppName,
							Label:   trigger.Label,
						}

						if trigger.AppName == "Shuffle Workflow" {
							foundAction.AppName = "shuffle-subflow"
						}

						break
					}
				}

				if len(foundAction.ID) == 0 {
					continue
				}
			}

			//log.Printf("\n\n\n[WARNING] Found that %s (%s) should be skipped? Should check if it has more parents. If not, send in a skip\n\n\n", foundAction.Label, foundAction.ID)

			foundCount := 0
			skippedBranches := []string{}
			for _, checkBranch := range workflowExecution.Workflow.Branches {
				if checkBranch.DestinationID == foundAction.ID {
					foundCount += 1

					// Check if they're all skipped or not
					if checkBranch.SourceID == actionResult.Action.ID {
						skippedBranches = append(skippedBranches, checkBranch.SourceID)
						continue
					}

					// Not found = not counted yet
					for _, res := range workflowExecution.Results {
						if res.Action.ID == checkBranch.SourceID && res.Status != "SUCCESS" && res.Status != "FINISHED" {
							skippedBranches = append(skippedBranches, checkBranch.SourceID)
							break
						}
					}
				}
			}

			skippedCount := len(skippedBranches)

			//log.Printf("\n\n[DEBUG][%s] Found %d branch(es) for %s. %d skipped. If equal, make the node skipped. SKIPPED: %s\n\n", workflowExecution.ExecutionId, foundCount, foundAction.Label, skippedCount, skippedBranches)
			if foundCount == skippedCount {
				found := false
				for _, res := range workflowExecution.Results {
					if res.Action.ID == foundAction.ID {
						found = true
					}
				}

				if !found {
					newResult := ActionResult{
						Action:        foundAction,
						ExecutionId:   actionResult.ExecutionId,
						Authorization: actionResult.Authorization,
						Result:        fmt.Sprintf(`{"success": false, "reason": "Skipped because of previous node (%s) - 1"}`, actionResult.Action.Label),
						StartedAt:     0,
						CompletedAt:   0,
						Status:        "SKIPPED",
					}

					resultData, err := json.Marshal(newResult)
					if err != nil {
						log.Printf("[ERROR] Failed skipping action")
						continue
					}

					streamUrl := fmt.Sprintf("http://localhost:5001/api/v1/streams")
					if project.Environment == "cloud" {
						streamUrl = fmt.Sprintf("https://shuffler.io/api/v1/streams")

						if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
							streamUrl = fmt.Sprintf("https://%s.%s.r.appspot.com/api/v1/streams", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
						}

						if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
							streamUrl = fmt.Sprintf("%s/api/v1/streams", os.Getenv("SHUFFLE_CLOUDRUN_URL"))
						}
					} else {
						if len(os.Getenv("WORKER_HOSTNAME")) > 0 {
							streamUrl = fmt.Sprintf("http://%s:33333/api/v1/streams", os.Getenv("WORKER_HOSTNAME"))
						}

						if os.Getenv("SHUFFLE_SWARM_CONFIG") == "run" && (project.Environment == "" || project.Environment == "worker") {

							streamUrl = fmt.Sprintf("http://localhost:33333/api/v1/streams")

						} else {
							if len(os.Getenv("BASE_URL")) > 0 {
								streamUrl = fmt.Sprintf("%s/api/v1/streams", os.Getenv("BASE_URL"))
							}
						}
					}

					req, err := http.NewRequest(
						"POST",
						streamUrl,
						bytes.NewBuffer([]byte(resultData)),
					)

					if err != nil {
						log.Printf("[ERROR] Error building SKIPPED request (%s): %s", foundAction.Label, err)
						continue
					}

					client := &http.Client{}
					newresp, err := client.Do(req)
					if err != nil {
						log.Printf("[ERROR] Error running SKIPPED request (%s): %s", foundAction.Label, err)
						continue
					}

					body, err := ioutil.ReadAll(newresp.Body)
					if err != nil {
						log.Printf("[ERROR] Failed reading body when running SKIPPED request (%s): %s", foundAction.Label, err)
						continue
					}

					//log.Printf("[DEBUG] Skipped body return from %s (%d): %s", streamUrl, newresp.StatusCode, string(body))
					if strings.Contains(string(body), "already finished") {
						log.Printf("[WARNING] Data couldn't be re-inputted for %s.", foundAction.Label)
						return &workflowExecution, true, errors.New(fmt.Sprintf("Failed updating skipped action %s", foundAction.Label))
					}
				}
			}
		}
	}

	// Related to notifications
	if actionResult.Status == "SUCCESS" && workflowExecution.Workflow.Configuration.SkipNotifications == false {
		// Marshal default failures
		resultCheck := ResultChecker{}
		err = json.Unmarshal([]byte(actionResult.Result), &resultCheck)
		if err == nil {
			//log.Printf("Unmarshal success!")
			if resultCheck.Success == false && strings.Contains(actionResult.Result, "success") && strings.Contains(actionResult.Result, "false") && workflowExecution.Workflow.Hidden == false {
				err = CreateOrgNotification(
					ctx,
					fmt.Sprintf("Potential error in Workflow %s", workflowExecution.Workflow.Name),
					fmt.Sprintf("Node %s in Workflow %s failed silently. Click to see more. Reason: %s", actionResult.Action.Label, workflowExecution.Workflow.Name, resultCheck.Reason),
					fmt.Sprintf("/workflows/%s?execution_id=%s&view=executions&node=%s", workflowExecution.Workflow.ID, workflowExecution.ExecutionId, actionResult.Action.ID),
					workflowExecution.ExecutionOrg,
					true,
				)

				if err != nil {
					log.Printf("[WARNING] Failed making org notification for %s: %s", workflowExecution.ExecutionOrg, err)
				}
			}
		} else {
			//log.Printf("[ERROR] Failed unmarshaling result into resultChecker (%s): %s", err, actionResult)
		}

		//log.Printf("[DEBUG] Ran marshal on silent failure")
	}

	// FIXME rebuild to be like this or something
	// workflowExecution/ExecutionId/Nodes/NodeId
	// Find the appropriate action
	if len(workflowExecution.Results) > 0 {
		// FIXME
		skip := false
		found := false
		outerindex := 0
		for index, item := range workflowExecution.Results {
			if item.Action.ID == actionResult.Action.ID {
				found = true
				if item.Status == actionResult.Status {
					skip = true
				}

				outerindex = index
				break
			}
		}

		if skip {
			//log.Printf("[DEBUG] Both results are %s. Skipping this node", item.Status)
		} else if found {
			// If result exists and execution variable exists, update execution value
			//log.Printf("Exec var backend: %s", workflowExecution.Results[outerindex].Action.ExecutionVariable.Name)
			actionVarName := workflowExecution.Results[outerindex].Action.ExecutionVariable.Name
			// Finds potential execution arguments
			if len(actionVarName) > 0 {
				//log.Printf("EXECUTION VARIABLE LOCAL: %s", actionVarName)
				for index, execvar := range workflowExecution.ExecutionVariables {
					if execvar.Name == actionVarName {
						// Sets the value for the variable

						if len(actionResult.Result) > 0 {
							log.Printf("\n\n[DEBUG] SET EXEC VAR\n\n", execvar.Name)
							workflowExecution.ExecutionVariables[index].Value = actionResult.Result
						} else {
							log.Printf("\n\n[DEBUG] SKIPPING EXEC VAR\n\n")
						}

						break
					}
				}
			}

			log.Printf("[INFO][%s] Updating %s in workflow from %s to %s", workflowExecution.ExecutionId, actionResult.Action.ID, workflowExecution.Results[outerindex].Status, actionResult.Status)

			actionResultBody, err := json.Marshal(actionResult)
			if err == nil {

				// Set cache for it too?
				cacheId := fmt.Sprintf("%s_%s_result", workflowExecution.ExecutionId, actionResult.Action.ID)
				err = SetCache(ctx, cacheId, actionResultBody, 35)
				if err != nil {
					log.Printf("[ERROR] Failed setting cache for User Input to %s: %s", actionResult.Status, err)
				} else {
					//log.Printf("[DEBUG] Set cache for SUBFLOW action result %s", cacheId)
				}
			} else {
				log.Printf("[ERROR] Failed marshaling action result for %s: %s", actionResult.Action.ID, err)
			}

			workflowExecution.Results[outerindex] = actionResult
		} else {
			//log.Printf("[INFO] Setting value of %s (%s) in workflow %s to %s (%d)", actionResult.Action.Label, actionResult.Action.ID, workflowExecution.ExecutionId, actionResult.Status, len(workflowExecution.Results))
			workflowExecution.Results = append(workflowExecution.Results, actionResult)
			//if subresult.Status == "SKIPPED" subresult.Status != "FAILURE" {
		}
	} else {
		log.Printf("[INFO] Setting value of %s (INIT - %s) in workflow %s to %s (%d)", actionResult.Action.Label, actionResult.Action.ID, workflowExecution.ExecutionId, actionResult.Status, len(workflowExecution.Results))
		workflowExecution.Results = append(workflowExecution.Results, actionResult)
	}

	// FIXME: Have a check for skippednodes and their parents
	/*
		for resultIndex, result := range workflowExecution.Results {
			if result.Status != "SKIPPED" {
				continue
			}

			// Checks if all parents are skipped or failed. Otherwise removes them from the results
			for _, branch := range workflowExecution.Workflow.Branches {
				if branch.DestinationID == result.Action.ID {
					for _, subresult := range workflowExecution.Results {
						if subresult.Action.ID == branch.SourceID {
							if subresult.Status != "SKIPPED" && subresult.Status != "FAILURE" {
								log.Printf("SUBRESULT PARENT STATUS: %s", subresult.Status)
								log.Printf("Should remove resultIndex: %d", resultIndex)

								workflowExecution.Results = append(workflowExecution.Results[:resultIndex], workflowExecution.Results[resultIndex+1:]...)

								break
							}
						}
					}
				}
			}
		}
	*/
	// Auto fixing and ensuring the same isn't ran multiple times?

	extraInputs := 0
	for _, trigger := range workflowExecution.Workflow.Triggers {
		if trigger.Name == "User Input" && trigger.AppName == "User Input" {
			extraInputs += 1
		} else if trigger.Name == "Shuffle Workflow" && trigger.AppName == "Shuffle Workflow" {
			extraInputs += 1
		}
	}

	//log.Printf("EXTRA: %d", extraInputs)
	//log.Printf("LENGTH: %d - %d", len(workflowExecution.Results), len(workflowExecution.Workflow.Actions)+extraInputs)
	updateParentRan := false
	if len(workflowExecution.Results) == len(workflowExecution.Workflow.Actions)+extraInputs {
		//log.Printf("\nIN HERE WITH RESULTS %d vs %d\n", len(workflowExecution.Results), len(workflowExecution.Workflow.Actions)+extraInputs)
		finished := true
		lastResult := ""

		// Doesn't have to be SUCCESS and FINISHED everywhere anymore.
		//skippedNodes := false
		for _, result := range workflowExecution.Results {
			if result.Status == "EXECUTING" || result.Status == "WAITING" {
				finished = false
				break
			}

			if result.Status == "SUCCESS" {
				lastResult = result.Result
			}
		}

		//log.Printf("[debug] Finished? %s", finished)
		if finished {
			dbSave = true
			if len(workflowExecution.ExecutionParent) == 0 {
				log.Printf("[INFO] Execution of %s in workflow %s finished (not subflow).", workflowExecution.ExecutionId, workflowExecution.Workflow.ID)
			} else {
				log.Printf("[INFO] SubExecution %s of parentExecution %s in workflow %s finished (subflow).", workflowExecution.ExecutionId, workflowExecution.ExecutionParent, workflowExecution.Workflow.ID)

			}

			for actionIndex, action := range workflowExecution.Workflow.Actions {
				for parameterIndex, param := range action.Parameters {
					if param.Configuration {
						//log.Printf("Cleaning up %s in %s", param.Name, action.Name)
						workflowExecution.Workflow.Actions[actionIndex].Parameters[parameterIndex].Value = ""
					}
				}
			}

			workflowExecution.Result = lastResult
			workflowExecution.Status = "FINISHED"
			workflowExecution.CompletedAt = int64(time.Now().Unix())
			if workflowExecution.LastNode == "" {
				workflowExecution.LastNode = actionResult.Action.ID
			}

			// 1. Check if the LAST node is FAILURE or ABORTED or SKIPPED
			// 2. If it's either of those, set the executionResult default value to DefaultReturnValue
			//log.Printf("\n\n===========\nSETTING VALUE TO %s\n============\nPARENT: %s\n\n", lastResult, workflowExecution.ExecutionParent)
			//log.Printf("\n\n===========\nSETTING VALUE TO %s\n============\nPARENT: %s\n\n", lastResult, workflowExecution.ExecutionParent)
			//log.Printf("%s", workflowExecution)

			valueToReturn := ""
			if len(workflowExecution.Workflow.DefaultReturnValue) > 0 {
				valueToReturn = workflowExecution.Workflow.DefaultReturnValue
				//log.Printf("\n\nCHECKING RESULT FOR LAST NODE %s with value \"%s\". Executionparent: %s\n\n", workflowExecution.ExecutionSourceNode, workflowExecution.Workflow.DefaultReturnValue, workflowExecution.ExecutionParent)
				for _, result := range workflowExecution.Results {
					if result.Action.ID == workflowExecution.LastNode {
						if result.Status == "ABORTED" || result.Status == "FAILURE" || result.Status == "SKIPPED" {
							workflowExecution.Result = workflowExecution.Workflow.DefaultReturnValue
							if len(workflowExecution.ExecutionParent) > 0 {
								// 1. Find the parent workflow
								// 2. Find the parent's existing value

								log.Printf("[DEBUG] FOUND SUBFLOW WITH EXECUTIONPARENT %s!", workflowExecution.ExecutionParent)
							}
						} else {
							valueToReturn = workflowExecution.Result
						}

						break
					}
				}
			} else {
				valueToReturn = workflowExecution.Result
			}

			// First: handle it in backend for loops
			// 2nd: Handle it in worker for normal executions
			/*
				if len(workflowExecution.ExecutionParent) > 0 && (project.Environment == "onprem") {
					log.Printf("[DEBUG] Got the result %s for subflow of %s. Check if this should be added to loop.", workflowExecution.Result, workflowExecution.ExecutionParent)

					parentExecution, err := GetWorkflowExecution(ctx, workflowExecution.ExecutionParent)
					if err == nil {
						isLooping := false
						for _, trigger := range parentExecution.Workflow.Triggers {
							if trigger.ID == workflowExecution.ExecutionSourceNode {
								for _, param := range trigger.Parameters {
									if param.Name == "argument" && strings.Contains(param.Value, "$") && strings.Contains(param.Value, ".#") {
										isLooping = true
										break
									}
								}

								break
							}
						}

						if isLooping {
							log.Printf("[DEBUG] Parentexecutions' subflow IS looping.")
						}
					}

				} else
			*/
			if len(workflowExecution.ExecutionParent) > 0 && len(workflowExecution.ExecutionSourceAuth) > 0 && len(workflowExecution.ExecutionSourceNode) > 0 {
				log.Printf("[DEBUG] Found execution parent %s for workflow '%s' (%s)", workflowExecution.ExecutionParent, workflowExecution.Workflow.Name, workflowExecution.Workflow.ID)

				err = updateExecutionParent(ctx, workflowExecution.ExecutionParent, valueToReturn, workflowExecution.ExecutionSourceAuth, workflowExecution.ExecutionSourceNode, workflowExecution.ExecutionId)
				if err != nil {
					log.Printf("[ERROR] Failed running update execution parent: %s", err)
				} else {
					updateParentRan = true
				}
			}
		}
	}

	// Had to move this to run AFTER "updateExecutionParent()", as it's controlling whether a subflow should be updated or not
	if actionResult.Status == "SUCCESS" && actionResult.Action.AppName == "shuffle-subflow" && !updateParentRan {
		runCheck := false
		for _, param := range actionResult.Action.Parameters {
			if param.Name == "check_result" {
				//log.Printf("[INFO] RESULT: %s", param)
				if param.Value == "true" {
					runCheck = true
				}

				break
			}
		}

		//if runCheck && project.Environment != "" && project.Environment != "worker" {
		if runCheck {
			// err = updateExecutionParent(workflowExecution.ExecutionParent, valueToReturn, workflowExecution.ExecutionSourceAuth, workflowExecution.ExecutionSourceNode)

			var subflowData SubflowData
			jsonerr := json.Unmarshal([]byte(actionResult.Result), &subflowData)

			// Big blob to check cache & backend for more results
			if jsonerr == nil && len(subflowData.Result) == 0 && !strings.Contains(actionResult.Result, "\"result\"") {
				if project.Environment != "cloud" {

					//Check cache for whether the execution actually finished or not
					// FIXMe: May need to get this from backend

					cacheKey := fmt.Sprintf("workflowexecution_%s", subflowData.ExecutionId)
					if value, found := requestCache.Get(cacheKey); found {
						parsedValue := WorkflowExecution{}
						cacheData := []byte(value.([]uint8))
						err = json.Unmarshal(cacheData, &parsedValue)
						if err == nil {
							log.Printf("[INFO][%s] Found subflow result (1) %s for subflow %s in recheck from cache with %d results and result %s", workflowExecution.ExecutionId, parsedValue.Status, subflowData.ExecutionId, len(parsedValue.Results), parsedValue.Result)

							if len(parsedValue.Result) > 0 {
								subflowData.Result = parsedValue.Result
							} else if parsedValue.Status == "FINISHED" {
								subflowData.Result = "Subflow finished (PS: This is from worker autofill - happens if no actual result in subflow exec)"
							}
						}

						// Check backend
						//log.Printf("[INFO][%s] Found subflow result %s for subflow %s in recheck from cache with %d results and result %s", workflowExecution.ExecutionId, parsedValue.Status, subflowData.ExecutionId, len(parsedValue.Results), parsedValue.Result)
						if len(subflowData.Result) == 0 && !strings.Contains(actionResult.Result, "\"result\"") {
							log.Printf("[INFO][%s] No subflow result found in cache for subflow %s. Checking backend next", workflowExecution.ExecutionId, subflowData.ExecutionId)
							if len(subflowData.ExecutionId) > 0 {
								parsedValue, err := GetBackendexecution(ctx, subflowData.ExecutionId, subflowData.Authorization)
								if err != nil {
									log.Printf("[WARNING] Failed getting subflow execution from backend to verify: %s", err)
								} else {
									log.Printf("[INFO][%s] Found subflow result (2) %s for subflow %s in backend with %d results and result %s", workflowExecution.ExecutionId, parsedValue.Status, subflowData.ExecutionId, len(parsedValue.Results), parsedValue.Result)
									if len(parsedValue.Result) > 0 {
										subflowData.Result = parsedValue.Result
									} else if parsedValue.Status == "FINISHED" {
										subflowData.Result = "Subflow finished (PS: This is from worker autofill - happens if no actual result in subflow exec)"
									}
								}
							}
						}
					}
				}
			}

			log.Printf("[WARNING][%s] Sinkholing request of %s IF the subflow-result DOESNT have result.", workflowExecution.ExecutionId, actionResult.Action.Label)
			if jsonerr == nil && len(subflowData.Result) == 0 && !strings.Contains(actionResult.Result, "\"result\"") {
				//func updateExecutionParent(executionParent, returnValue, parentAuth, parentNode string) error {
				log.Printf("[INFO][%s] NO RESULT FOR SUBFLOW RESULT - SETTING TO EXECUTING. Results: %d. Trying to find subexec in cache onprem\n\n", workflowExecution.ExecutionId, len(workflowExecution.Results))

				// Finding the result, and removing it if it exists. "Sinkholing"
				workflowExecution.Status = "EXECUTING"
				newResults := []ActionResult{}
				for _, result := range workflowExecution.Results {
					if result.Action.ID == actionResult.Action.ID {
						continue
					}

					newResults = append(newResults, result)
				}

				workflowExecution.Results = newResults

				//for _, result := range
			} else {
				var subflowDataList []SubflowData
				err = json.Unmarshal([]byte(actionResult.Result), &subflowDataList)
				if err != nil || len(subflowDataList) == 0 {
					log.Printf("\n\nNOT sinkholed: %s", err)
					for resultIndex, result := range workflowExecution.Results {
						if result.Action.ID == actionResult.Action.ID {
							workflowExecution.Results[resultIndex] = actionResult
							break
						}
					}

				} else {
					log.Printf("[WARNING] LIST sinkholed (%d) - Should apply list setup for same as subflow without result! Set the execution back to EXECUTING and the action to WAITING, as it's already running. Waiting for each individual result to add to the list.", len(subflowDataList))

					// Set to executing, as the point is for the subflows themselves to update this part. This does NOT happen in the subflow, but in the parent workflow, which is waiting for results to be ingested, hence it's set to EXECUTING
					workflowExecution.Status = "EXECUTING"

					// Setting to waiting, as it should be updated by child executions
					actionResult.Status = "WAITING"
					for resultIndex, result := range workflowExecution.Results {
						if result.Action.ID == actionResult.Action.ID {
							workflowExecution.Results[resultIndex] = actionResult

							actionResultBody, err := json.Marshal(actionResult)
							if err == nil {
								cacheId := fmt.Sprintf("%s_%s_result", workflowExecution.ExecutionId, actionResult.Action.ID)
								err = SetCache(ctx, cacheId, actionResultBody, 35)
								if err != nil {
									log.Printf("[ERROR] Failed setting cache for SUBFLOW to WAITING: %s", err)
								} else {
									//log.Printf("[DEBUG] Set cache for SUBFLOW action result %s", cacheId)
								}
							}

							break
						}
					}

					/*
						for _, subflowItem := range subflowDataList {
							log.Printf("%s == %s", subflowItem.ExecutionId, workflowExecution.ExecutionId)

							if len(subflowItem.Result) == 0 {
								subflowItem.Result = workflowExecution.Result

								//if subflowItem.ExecutionId == workflowExecution.ExecutionId {
								//	log.Printf("FOUND EXECUTION ID IN SUBFLOW: %s", subflowItem.ExecutionId)
								//tmpJson, err := json.Marshal(workflowExecution)
								//if strings.Contains(
							}
						}
					*/
				}

				dbSave = true
			}
		}
	}

	workflowExecution, newDbSave := compressExecution(ctx, workflowExecution, "mid-cleanup")
	if !dbSave {
		dbSave = newDbSave
	}

	// Does it work to cache it here?
	err = SetWorkflowExecution(ctx, workflowExecution, dbSave)
	if err != nil {
		log.Printf("[ERROR][%s] Failed saving execution to DB: %s", workflowExecution.ExecutionId, err)
	}

	// Should only apply a few seconds after execution, otherwise it's bascially spam.
	if !skipExecutionCount && workflowExecution.Status == "FINISHED" {
		IncrementCache(ctx, workflowExecution.ExecutionOrg, "workflow_executions_finished")
	}

	// Should this be able to return errors?
	//return &workflowExecution, dbSave, err
	return &workflowExecution, dbSave, nil
}

// Finds execution results and parameters that are too large to manage and reduces them / saves data partly
func compressExecution(ctx context.Context, workflowExecution WorkflowExecution, saveLocationInfo string) (WorkflowExecution, bool) {

	//GetApp(ctx context.Context, id string, user User) (*WorkflowApp, error) {
	//return workflowExecution, false
	dbSave := false
	tmpJson, err := json.Marshal(workflowExecution)
	if err == nil {
		if project.DbType != "opensearch" {
			if len(tmpJson) >= 1000000 {
				// Clean up results' actions

				dbSave = true
				//log.Printf("[WARNING] Result length is too long (%d) when running %s! Need to reduce result size. Attempting auto-compression by saving data to disk.", len(tmpJson), saveLocationInfo)
				actionId := "execution_argument"

				//gs://shuffler.appspot.com/extra_specs/0373ed696a3a2cba0a2b6838068f2b80
				//log.Printf("[WARNING] Couldn't find  for %s. Should check filepath gs://%s/%s (size too big)", innerApp.ID, internalBucket, fullParsedPath)

				// Result        string `json:"result" datastore:"result,noindex"`
				// Arbitrary reduction size
				maxSize := 50000
				bucketName := "shuffler.appspot.com"

				if len(workflowExecution.ExecutionArgument) > maxSize {
					itemSize := len(workflowExecution.ExecutionArgument)
					baseResult := fmt.Sprintf(`{
								"success": False,
								"reason": "Result too large to handle (https://github.com/frikky/shuffle/issues/171)."
								"size": %d,
								"extra": "",
								"id": "%s_%s"
							}`, itemSize, workflowExecution.ExecutionId, actionId)

					fullParsedPath := fmt.Sprintf("large_executions/%s/%s_%s", workflowExecution.ExecutionOrg, workflowExecution.ExecutionId, actionId)
					log.Printf("[DEBUG] Saving value of %s to storage path %s", actionId, fullParsedPath)
					bucket := project.StorageClient.Bucket(bucketName)
					obj := bucket.Object(fullParsedPath)
					w := obj.NewWriter(ctx)
					if _, err := fmt.Fprint(w, workflowExecution.ExecutionArgument); err != nil {
						log.Printf("[WARNING] Failed writing new exec file: %s", err)
						workflowExecution.ExecutionArgument = baseResult
						//continue
					} else {
						// Close, just like writing a file.
						if err := w.Close(); err != nil {
							log.Printf("[WARNING] Failed closing new exec file: %s", err)
							workflowExecution.ExecutionArgument = baseResult
						} else {
							workflowExecution.ExecutionArgument = fmt.Sprintf(`{
								"success": False,
								"reason": "Result too large to handle (https://github.com/frikky/shuffle/issues/171).",
								"size": %d,
								"extra": "replace",
								"id": "%s_%s"
							}`, itemSize, workflowExecution.ExecutionId, actionId)
						}
					}
				}

				newResults := []ActionResult{}
				//shuffle-large-executions
				for _, item := range workflowExecution.Results {
					if len(item.Result) > maxSize {

						itemSize := len(item.Result)
						baseResult := fmt.Sprintf(`{
								"success": False,
								"reason": "Result too large to handle (https://github.com/frikky/shuffle/issues/171)."
								"size": %d,
								"extra": "",
								"id": "%s_%s"
							}`, itemSize, workflowExecution.ExecutionId, item.Action.ID)

						// 1. Get the value and set it instead if it exists
						// 2. If it doesn't exist, add it
						_, err := getExecutionFileValue(ctx, workflowExecution, item)
						if err == nil {
							//log.Printf("[DEBUG] Found execution locally for %s. Not saving another.", item.Action.Label)
						} else {
							fullParsedPath := fmt.Sprintf("large_executions/%s/%s_%s", workflowExecution.ExecutionOrg, workflowExecution.ExecutionId, item.Action.ID)
							log.Printf("[DEBUG] Saving value of %s to storage path %s", item.Action.ID, fullParsedPath)
							bucket := project.StorageClient.Bucket(bucketName)
							obj := bucket.Object(fullParsedPath)
							w := obj.NewWriter(ctx)
							//log.Printf("RES: ", item.Result)
							if _, err := fmt.Fprint(w, item.Result); err != nil {
								log.Printf("[WARNING] Failed writing new exec file: %s", err)
								item.Result = baseResult
								newResults = append(newResults, item)
								continue
							}

							// Close, just like writing a file.
							if err := w.Close(); err != nil {
								log.Printf("[WARNING] Failed closing new exec file: %s", err)
								item.Result = baseResult
								newResults = append(newResults, item)
								continue
							}
						}

						item.Result = fmt.Sprintf(`{
								"success": False,
								"reason": "Result too large to handle (https://github.com/frikky/shuffle/issues/171).",
								"size": %d,
								"extra": "replace",
								"id": "%s_%s"
							}`, itemSize, workflowExecution.ExecutionId, item.Action.ID)
						// Setting an arbitrary decisionpoint to get it
						// Backend will use this ID + action ID to get the data back
						//item.Result = fmt.Sprintf("EXECUTION=%s", workflowExecution.ExecutionId)
					}

					newResults = append(newResults, item)
				}

				workflowExecution.Results = newResults
			}

			jsonString, err := json.Marshal(workflowExecution)
			if err == nil {
				//log.Printf("Execution size: %d", len(jsonString))
				if len(jsonString) > 1000000 {
					//for _, action := range workflowExecution.Workflow.Actions {
					//	actionData, err := json.Marshal(action)
					//	if err == nil {
					//		//log.Printf("[DEBUG] Action Size for %s (%s - %s): %d", action.Label, action.Name, action.ID, len(actionData))
					//	}
					//}

					for resultIndex, result := range workflowExecution.Results {
						//resultData, err := json.Marshal(result)
						//_ = resultData
						actionData, err := json.Marshal(result.Action)
						if err == nil {
							//log.Printf("Result Size (%s - action: %d): %d. Value size: %d", result.Action.Label, len(resultData), len(actionData), len(result.Result))
						}

						if len(actionData) > 10000 {
							for paramIndex, param := range result.Action.Parameters {
								if len(param.Value) > 10000 {
									workflowExecution.Results[resultIndex].Action.Parameters[paramIndex].Value = "Size too large. Removed."
								}
							}
						}
					}
				}
			}
		}
	}

	return workflowExecution, dbSave
}

// Finds the child nodes of a node in execution and returns them
// Used if e.g. a node in a branch is exited, and all children have to be stopped
func FindChildNodes(workflowExecution WorkflowExecution, nodeId string, parents, handledBranches []string) []string {
	//log.Printf("\nNODE TO FIX: %s\n\n", nodeId)
	allChildren := []string{nodeId}
	//log.Printf("\n\n")

	// 1. Find children of this specific node
	// 2. Find the children of those nodes etc.
	// 3. Sort it in the right order to handle merges properly
	for _, branch := range workflowExecution.Workflow.Branches {
		if branch.SourceID == nodeId {
			if ArrayContains(parents, branch.DestinationID) {
				continue
			}

			parents = append(parents, branch.SourceID)
			if ArrayContains(handledBranches, branch.ID) {
				continue
			}

			//log.Printf("NODE: %s, SRC: %s, CHILD: %s\n", nodeId, branch.SourceID, branch.DestinationID)
			allChildren = append(allChildren, branch.DestinationID)

			handledBranches = append(handledBranches, branch.ID)
			childNodes := FindChildNodes(workflowExecution, branch.DestinationID, parents, handledBranches)
			for _, bottomChild := range childNodes {
				found := false

				for _, topChild := range allChildren {
					if topChild == bottomChild {
						found = true
						break
					}
				}

				if !found {
					allChildren = append(allChildren, bottomChild)
				}
			}
		}
	}

	// Remove potential duplicates
	newNodes := []string{}
	for _, tmpnode := range allChildren {
		if tmpnode == nodeId {
			continue
		}

		found := false
		for _, newnode := range newNodes {
			if newnode == tmpnode {
				found = true
				break
			}
		}

		if !found {
			newNodes = append(newNodes, tmpnode)
		}
	}

	return newNodes
}

func ActivateWorkflowApp(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get active apps: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to activate workflow app (shared): %s (%s)", user.Username, user.Id)
		resp.WriteHeader(403)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	ctx := GetContext(request)
	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	app, err := GetApp(ctx, fileId, user, false)
	if err != nil {
		appName := request.URL.Query().Get("app_name")
		appVersion := request.URL.Query().Get("app_version")

		if len(appName) > 0 && len(appVersion) > 0 {
			apps, err := FindWorkflowAppByName(ctx, appName)
			//log.Printf("[INFO] Found %d apps for %s", len(apps), appName)
			if err != nil || len(apps) == 0 {
				log.Printf("[WARNING] Error getting app %s (app config): %s", appName, err)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false, "reason": "App doesn't exist"}`))
				return
			}

			selectedApp := WorkflowApp{}
			for _, app := range apps {
				if !app.Sharing && !app.Public {
					continue
				}

				if app.Name == appName {
					selectedApp = app
				}

				if app.Name == appName && app.AppVersion == appVersion {
					selectedApp = app
				}
			}

			app = &selectedApp
		} else {
			log.Printf("[WARNING] Error getting app with ID %s (app config): %s", fileId, err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "App doesn't exist"}`))
			return
		}
	}

	org := &Org{}
	added := false
	if app.Sharing || app.Public {
		org, err = GetOrg(ctx, user.ActiveOrg.Id)
		if err == nil {
			if len(org.ActiveApps) > 150 {
				// No reason for it to be this big. Arbitrarily reducing.
				same := []string{}
				samecnt := 0
				for _, activeApp := range org.ActiveApps {
					if ArrayContains(same, activeApp) {
						samecnt += 1
						continue
					}

					same = append(same, activeApp)
				}

				added = true
				//log.Printf("Same: %d, total uniq: %d", samecnt, len(same))
				org.ActiveApps = org.ActiveApps[len(org.ActiveApps)-100 : len(org.ActiveApps)-1]
			}

			if !ArrayContains(org.ActiveApps, app.ID) {
				org.ActiveApps = append(org.ActiveApps, app.ID)
				added = true
			}

			if added {
				err = SetOrg(ctx, *org, org.Id)
				if err != nil {
					log.Printf("[WARNING] Failed setting org when autoadding apps on save: %s", err)
				} else {
					log.Printf("[INFO] Added public app %s (%s) to org %s (%s)", app.Name, app.ID, user.ActiveOrg.Name, user.ActiveOrg.Id)
					DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
					DeleteCache(ctx, fmt.Sprintf("apps_%s", user.ActiveOrg.Id))
				}
			}
		}
	} else {
		log.Printf("[WARNING] User is trying to activate %s which is NOT a public app", app.Name)
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	log.Printf("[DEBUG] App %s (%s) activated for org %s by user %s (%s). Active apps: %d. Already existed: %t", app.Name, app.ID, user.ActiveOrg.Id, user.Username, user.Id, len(org.ActiveApps), !added)
	DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
	DeleteCache(ctx, fmt.Sprintf("apps_%s", user.ActiveOrg.Id))
	DeleteCache(ctx, "all_apps")
	DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-100"))
	DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-500"))
	DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-1000"))

	// If onprem, it should autobuild the container(s) from here

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

func GetExecutionbody(body []byte) string {
	parsedBody := string(body)

	// Specific weird newline issues
	if strings.Contains(parsedBody, "choice") {
		if strings.Count(parsedBody, `\\n`) > 2 {
			parsedBody = strings.Replace(parsedBody, `\\n`, "", -1)
		}
		if strings.Count(parsedBody, `\u0022`) > 2 {
			parsedBody = strings.Replace(parsedBody, `\u0022`, `"`, -1)
		}
		if strings.Count(parsedBody, `\\"`) > 2 {
			parsedBody = strings.Replace(parsedBody, `\\"`, `"`, -1)
		}

		if strings.Contains(parsedBody, `"extra": "{`) {
			parsedBody = strings.Replace(parsedBody, `"extra": "{`, `"extra": {`, 1)
			parsedBody = strings.Replace(parsedBody, `}"}`, `}}`, 1)
		}
	}

	// Replaces dots in string when it's key specifically has a dot
	// FIXME: Do this with key recursion and key replacements only
	pattern := regexp.MustCompile(`\"(\w+)\.(\w+)\":`)
	found := pattern.FindAllString(parsedBody, -1)
	for _, item := range found {
		newItem := strings.Replace(item, ".", "_", -1)
		parsedBody = strings.Replace(parsedBody, item, newItem, -1)
	}

	if !strings.HasPrefix(parsedBody, "{") && !strings.HasPrefix(parsedBody, "[") && strings.Contains(parsedBody, "=") {
		log.Printf("[DEBUG] Trying to make string %s to json (skipping if XML)", parsedBody)

		// Dumb XML handler
		if strings.HasPrefix(strings.TrimSpace(parsedBody), "<") && strings.HasSuffix(strings.TrimSpace(parsedBody), ">") {
			log.Printf("[DEBUG] XML detected. Not parsing anyything.")
			return parsedBody
		}

		newbody := map[string]string{}
		for _, item := range strings.Split(parsedBody, "&") {
			//log.Printf("Handling item: %s", item)

			if !strings.Contains(item, "=") {
				newbody[item] = ""
				continue
			}

			bodySplit := strings.Split(item, "=")
			if len(bodySplit) == 2 {
				newbody[bodySplit[0]] = bodySplit[1]
			} else {
				newbody[item] = ""
			}
		}

		jsonString, err := json.Marshal(newbody)
		if err != nil {
			log.Printf("[ERROR] Failed marshaling queries: %s: %s", newbody, err)
		} else {
			parsedBody = string(jsonString)
		}
		//fmt.Printf(err)
		//log.Printf("BODY: %s", newbody)
	}

	// Check bad characters in keys
	// FIXME: Re-enable this when it's safe.
	//log.Printf("Input: %s", parsedBody)
	parsedBody = string(FixBadJsonBody([]byte(parsedBody)))
	//log.Printf("Output: %s", parsedBody)

	return parsedBody
}

// Can't just regex out stuff due to unicode problems with other languages
func handleKeyRemoval(key string) string {
	abolish := []string{"!", "@", "#", "$", "%", "~", "|", "^", "&", "*", "(", ")", "[", "]", "{", "}", "<", ">", "+", "=", "?", ".", ",", "/", "\\", "'"}

	for _, remove := range abolish {
		key = strings.Replace(key, remove, "", -1)
	}

	return key
}

// https://www.codemio.com/2021/02/advanced-golang-tutorials-dynamic-json-parsing.html
func handleJSONObject(object interface{}, key, totalObject string) string {
	currentObject := ""
	key = handleKeyRemoval(key)

	switch t := object.(type) {
	case int:
		currentObject += fmt.Sprintf(`"%s": %d, `, key, t)
		if len(key) == 0 {
			currentObject += fmt.Sprintf(`%d, `, t)
		}
	case int64:
		currentObject += fmt.Sprintf(`"%s": %d, `, key, t)
		if len(key) == 0 {
			currentObject += fmt.Sprintf(`%d, `, t)
		}
	case float64:
		tmpObject := fmt.Sprintf(`"%s": %f, `, key, t)
		if len(key) == 0 {
			tmpObject = fmt.Sprintf(`%f, `, t)
		}

		if strings.HasSuffix(tmpObject, "000000, ") {
			tmpObject = tmpObject[0 : len(tmpObject)-9]
			tmpObject += ", "
		}

		currentObject += tmpObject
	case bool:
		if len(key) == 0 {
			currentObject += fmt.Sprintf(`%v, `, t)
		} else {
			currentObject += fmt.Sprintf(`"%s": %v, `, key, t)
		}
	case string:
		if len(key) == 0 {
			currentObject += fmt.Sprintf(`"%s", `, t)
		} else {
			currentObject += fmt.Sprintf(`"%s": "%s", `, key, t)
		}
	case map[string]interface{}:
		if len(key) == 0 {
			currentObject += fmt.Sprintf(`{`)
		} else {
			currentObject += fmt.Sprintf(`"%s": {`, key)
		}

		for k, v := range t {
			currentObject = handleJSONObject(v, k, currentObject)
		}

		if len(currentObject) > 3 {
			currentObject = currentObject[0 : len(currentObject)-2]
		}

		currentObject += "}, "
	case []interface{}:
		if len(key) == 0 {
			currentObject += fmt.Sprintf(`[`)
		} else {
			currentObject += fmt.Sprintf(`"%s": [`, key)
		}

		for _, v := range t {
			currentObject = handleJSONObject(v, "", currentObject)
		}

		if len(currentObject) > 3 {
			currentObject = currentObject[0 : len(currentObject)-2]
		}

		currentObject += "], "
	default:
		log.Printf("[ERROR] Missing handler for type %s in app framework - key: %s", t, key)
	}

	totalObject += currentObject
	return totalObject
}

func FixBadJsonBody(parsedBody []byte) []byte {
	if os.Getenv("SHUFFLE_JSON_PARSER") != "parse" {
		return parsedBody
	}
	// NOT handling data that starts as a loop for now: [] instead of {} as outer wrapper.
	// Lists and all other types do work inside the JSON, and are rebuilt with a new key (if applicable).

	if !strings.HasPrefix(string(parsedBody), "{") {
		return parsedBody
	}

	var results map[string]interface{}
	err := json.Unmarshal([]byte(parsedBody), &results)
	if err != nil {
		log.Printf("[WARNING] Failed parsing data: %s", err)
		return parsedBody
	}

	totalObject := "{"
	for key, value := range results {
		_ = value
		totalObject = handleJSONObject(value, key, totalObject)
	}

	if len(totalObject) > 3 {
		totalObject = totalObject[0 : len(totalObject)-2]
	}

	totalObject += "}"

	//log.Printf("Auto sanitized keys.: %s", totalObject)
	//for _, result := range results {
	//	// But if you don't know the field types, you can use type switching to determine (safe):
	//	// Keep in mind that, since this is a map, the order is not guaranteed.
	//	fmt.Printf("\nType Switching: ")
	//	for k := range result {
	//	}

	//	fmt.Printf("------------------------------")
	//}

	return []byte(totalObject)
}

func ValidateSwagger(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Just here to verify that the user is logged in
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in validate swagger: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to validate swagger (shared): %s (%s)", user.Username, user.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed reading body"}`))
		return
	}

	type versionCheck struct {
		Swagger        string `datastore:"swagger" json:"swagger" yaml:"swagger"`
		SwaggerVersion string `datastore:"swaggerVersion" json:"swaggerVersion" yaml:"swaggerVersion"`
		OpenAPI        string `datastore:"openapi" json:"openapi" yaml:"openapi"`
	}

	//body = []byte(`swagger: "2.0"`)
	//body = []byte(`swagger: '1.0'`)
	//newbody := string(body)
	//newbody = strings.TrimSpace(newbody)
	//body = []byte(newbody)
	//log.Printf(string(body))
	//tmpbody, err := yaml.YAMLToJSON(body)
	//log.Printf(err)
	//log.Printf(string(tmpbody))

	// This has to be done in a weird way because Datastore doesn't
	// support map[string]interface and similar (openapi3.Swagger)
	var version versionCheck

	re := regexp.MustCompile("[[:^ascii:]]")
	//re := regexp.MustCompile("[[:^unicode:]]")
	t := re.ReplaceAllLiteralString(string(body), "")
	log.Printf("[DEBUG] App build API length: %d. Cleanup length: %d", len(string(body)), len(t))
	body = []byte(t)

	isJson := false
	err = json.Unmarshal(body, &version)
	if err != nil {
		log.Printf("[WARNING] Json upload err: %s", err)

		body = []byte(strings.Replace(string(body), "\\/", "/", -1))
		err = yaml.Unmarshal(body, &version)
		if err != nil {
			log.Printf("[WARNING] Yaml error (3): %s", err)
			//if len(string(body)) < 500 {
			//	log.Printf("%s",
			//}
			resp.WriteHeader(422)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed reading openapi to json and yaml. Is version defined?: %s"}`, err)))
			return
		} else {
			log.Printf("[INFO] Successfully parsed YAML (3)!")
		}
	} else {
		isJson = true
		log.Printf("[INFO] Successfully parsed JSON!")
	}

	if len(version.SwaggerVersion) > 0 && len(version.Swagger) == 0 {
		version.Swagger = version.SwaggerVersion
	}
	log.Printf("[INFO] Version: %s", version)
	log.Printf("[INFO] OpenAPI: %s", version.OpenAPI)

	if strings.HasPrefix(version.Swagger, "3.") || strings.HasPrefix(version.OpenAPI, "3.") {
		log.Printf("[INFO] Handling v3 API")
		swaggerLoader := openapi3.NewSwaggerLoader()
		swaggerLoader.IsExternalRefsAllowed = true
		swagger, err := swaggerLoader.LoadSwaggerFromData(body)
		if err != nil {
			log.Printf("[WARNING] Failed to convert v3 API: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, err)))
			return
		}

		hasher := md5.New()
		hasher.Write(body)
		idstring := hex.EncodeToString(hasher.Sum(nil))

		log.Printf("[INFO] Swagger v3 validation success with ID %s and %d paths!", idstring, len(swagger.Paths))

		if !isJson {
			log.Printf("[INFO] NEED TO TRANSFORM FROM YAML TO JSON for %s", idstring)
		}

		swaggerdata, err := json.Marshal(swagger)
		if err != nil {
			log.Printf("[WARNING] Failed unmarshaling v3 data: %s", err)
			resp.WriteHeader(422)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed marshalling swaggerv3 data: %s"}`, err)))
			return
		}
		parsed := ParsedOpenApi{
			ID:   idstring,
			Body: string(swaggerdata),
		}

		ctx := GetContext(request)
		err = SetOpenApiDatastore(ctx, idstring, parsed)
		if err != nil {
			log.Printf("[WARNING] Failed uploading openapi to datastore: %s", err)
			resp.WriteHeader(422)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed reading openapi2: %s"}`, err)))
			return
		}

		log.Printf("[INFO] Successfully set OpenAPI with ID %s", idstring)
		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf(`{"success": true, "id": "%s"}`, idstring)))
		return
	} else { //strings.HasPrefix(version.Swagger, "2.") || strings.HasPrefix(version.OpenAPI, "2.") {
		// Convert
		log.Printf("[WARNING] Handling v2 API")
		swagger := openapi2.Swagger{}
		//log.Printf(string(body))
		err = json.Unmarshal(body, &swagger)
		if err != nil {
			log.Printf("[WARNING] Json error for v2 - trying yaml next: %s", err)
			err = yaml.Unmarshal([]byte(body), &swagger)
			if err != nil {
				log.Printf("[WARNING] Yaml error (4): %s", err)

				resp.WriteHeader(422)
				resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed reading openapi2: %s"}`, err)))
				return
			} else {
				log.Printf("Found valid yaml!")
			}

		}

		swaggerv3, err := openapi2conv.ToV3Swagger(&swagger)
		if err != nil {
			log.Printf("[WARNING] Failed converting from openapi2 to 3: %s", err)
			resp.WriteHeader(422)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed converting from openapi2 to openapi3: %s"}`, err)))
			return
		}

		swaggerdata, err := json.Marshal(swaggerv3)
		if err != nil {
			log.Printf("[WARNING] Failed unmarshaling v3 from v2 data: %s", err)
			resp.WriteHeader(422)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed marshalling swaggerv3 data: %s"}`, err)))
			return
		}

		hasher := md5.New()
		hasher.Write(swaggerdata)
		idstring := hex.EncodeToString(hasher.Sum(nil))
		if !isJson {
			log.Printf("[WARNING] FIXME: NEED TO TRANSFORM FROM YAML TO JSON for %s?", idstring)
		}
		log.Printf("[INFO] Swagger v2 -> v3 validation success with ID %s!", idstring)

		parsed := ParsedOpenApi{
			ID:   idstring,
			Body: string(swaggerdata),
		}

		ctx := GetContext(request)
		err = SetOpenApiDatastore(ctx, idstring, parsed)
		if err != nil {
			log.Printf("[WARNING] Failed uploading openapi2 to datastore: %s", err)
			resp.WriteHeader(422)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed reading openapi2: %s"}`, err)))
			return
		}

		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf(`{"success": true, "id": "%s"}`, idstring)))
		return
	}
	/*
		else {
			log.Printf("Swagger / OpenAPI version %s is not supported or there is an error.", version.Swagger)
			resp.WriteHeader(422)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Swagger version %s is not currently supported"}`, version.Swagger)))
			return
		}
	*/

	// save the openapi ID
	resp.WriteHeader(422)
	resp.Write([]byte(`{"success": false}`))
}

// Recursively finds child nodes inside sub workflows
func GetReplacementNodes(ctx context.Context, execution WorkflowExecution, trigger Trigger, lastTriggerName string) ([]Action, []Branch, string) {
	if execution.ExecutionOrg == "" {
		execution.ExecutionOrg = execution.Workflow.OrgId
	}

	selectedWorkflow := ""
	workflowAction := ""
	for _, param := range trigger.Parameters {
		if param.Name == "workflow" {
			selectedWorkflow = param.Value
		}

		if param.Name == "startnode" {
			workflowAction = param.Value
		}
	}

	if len(selectedWorkflow) == 0 {
		return []Action{}, []Branch{}, ""
	}

	// Authenticating and such
	workflow, err := GetWorkflow(ctx, selectedWorkflow)
	if err != nil {
		return []Action{}, []Branch{}, ""
	}

	orgFound := false
	if workflow.ExecutingOrg.Id == execution.ExecutionOrg {
		orgFound = true
	} else if workflow.OrgId == execution.ExecutionOrg {
		orgFound = true
	} else {
		for _, org := range workflow.Org {
			if org.Id == execution.ExecutionOrg {
				orgFound = true
				break
			}
		}
	}

	if !orgFound {
		log.Printf("[WARNING] Auth for subflow is bad. %s (orig) vs %s", execution.ExecutionOrg, workflow.OrgId)
		return []Action{}, []Branch{}, ""
	}

	//childNodes = FindChildNodes(workflowExecution, actionResult.Action.ID)
	//log.Printf("FIND CHILDNODES OF STARTNODE %s", workflowAction)
	workflowExecution := WorkflowExecution{
		Workflow: *workflow,
	}

	childNodes := FindChildNodes(workflowExecution, workflowAction, []string{}, []string{})
	//log.Printf("Found %d childnodes of %s", len(childNodes), workflowAction)
	newActions := []Action{}
	branches := []Branch{}

	// FIXME: Bad lastnode check. Need to go to the bottom of workflows and check max steps away from parent
	lastNode := ""
	for _, nodeId := range childNodes {
		for _, action := range workflow.Actions {
			if nodeId == action.ID {
				newActions = append(newActions, action)
				break
			}
		}

		for _, branch := range workflow.Branches {
			if branch.SourceID == nodeId {
				branches = append(branches, branch)
			}
		}

		lastNode = nodeId
	}

	found := false
	for actionIndex, action := range newActions {
		if lastNode == action.ID {
			//actions[actionIndex].Name = trigger.Name
			newActions[actionIndex].Label = lastTriggerName
			//trigger.Label
			found = true
		}
	}

	if !found {
		log.Printf("SHOULD CHECK TRIGGERS FOR LASTNODE!")
	}

	log.Printf("[INFO] Found %d actions and %d branches in subflow", len(newActions), len(branches))
	if len(newActions) == len(childNodes) {
		return newActions, branches, lastNode
	} else {
		log.Printf("\n\n[WARNING] Bad length of actions and nodes in subflow (subsubflow?): %d vs %d", len(newActions), len(childNodes))

		// Adding information about triggers if subflow
		changed := false
		for _, nodeId := range childNodes {
			for triggerIndex, trigger := range workflow.Triggers {
				if trigger.AppName == "Shuffle Workflow" {
					if nodeId == trigger.ID {
						replaceActions := false
						workflowAction := ""
						for _, param := range trigger.Parameters {
							if param.Name == "argument" && !strings.Contains(param.Value, ".#") {
								replaceActions = true
							}

							if param.Name == "startnode" {
								workflowAction = param.Value
							}
						}

						if replaceActions {
							replacementNodes, newBranches, lastNode := GetReplacementNodes(ctx, workflowExecution, trigger, lastTriggerName)
							log.Printf("SUB REPLACEMENTS: %d, %d", len(replacementNodes), len(newBranches))
							log.Printf("\n\nNEW LASTNODE: %s\n\n", lastNode)
							if len(replacementNodes) > 0 {
								//workflowExecution.Workflow.Actions = append(workflowExecution.Workflow.Actions, action)

								//lastnode = replacementNodes[0]
								// Have to validate in case it's the same workflow and such
								for _, action := range replacementNodes {
									found := false
									for subActionIndex, subaction := range newActions {
										if subaction.ID == action.ID {
											found = true
											//newActions[subActionIndex].Name = action.Name
											newActions[subActionIndex].Label = action.Label
											break
										}
									}

									if !found {
										action.SubAction = true
										newActions = append(newActions, action)
									}
								}

								for _, branch := range newBranches {
									workflowExecution.Workflow.Branches = append(workflowExecution.Workflow.Branches, branch)
								}

								// Append branches:
								// parent -> new inner node (FIRST one)
								for branchIndex, branch := range workflowExecution.Workflow.Branches {
									if branch.DestinationID == trigger.ID {
										log.Printf("REPLACE DESTINATION WITH %s!!", workflowAction)
										workflowExecution.Workflow.Branches[branchIndex].DestinationID = workflowAction
										branches = append(branches, workflowExecution.Workflow.Branches[branchIndex])
									}

									if branch.SourceID == trigger.ID {
										log.Printf("REPLACE SOURCE WITH LASTNODE %s!!", lastNode)
										workflowExecution.Workflow.Branches[branchIndex].SourceID = lastNode
										branches = append(branches, workflowExecution.Workflow.Branches[branchIndex])
									}
								}

								// Remove the trigger
								workflowExecution.Workflow.Triggers = append(workflowExecution.Workflow.Triggers[:triggerIndex], workflowExecution.Workflow.Triggers[triggerIndex+1:]...)
								workflow.Triggers = append(workflow.Triggers[:triggerIndex], workflow.Triggers[triggerIndex+1:]...)
								changed = true
							}
						}
					}
				}

			}
		}

		log.Printf("NEW ACTION LENGTH %d, Branches: %d. LASTNODE: %s\n\n", len(newActions), len(branches), lastNode)
		if changed {
			return newActions, branches, lastNode
		}
	}

	return []Action{}, []Branch{}, ""
}

// Uses a simple way to be able to modify the encryption key being used
// FIXME: Investigate better ways of handling EVERYTHING related to encryption
// E.g. rolling keys and such
func create32Hash(key string) ([]byte, error) {
	encryptionModifier := os.Getenv("SHUFFLE_ENCRYPTION_MODIFIER")
	if len(encryptionModifier) == 0 {
		return []byte{}, errors.New(fmt.Sprintf("No encryption modifier set. Define SHUFFLE_ENCRYPTION_MODIFIER and NEVER change it to start encrypting auth."))
	}

	key += encryptionModifier
	hasher := md5.New()
	hasher.Write([]byte(key))
	return []byte(hex.EncodeToString(hasher.Sum(nil))), nil
}

func handleKeyEncryption(data []byte, passphrase string) ([]byte, error) {
	key, err := create32Hash(passphrase)
	if err != nil {
		log.Printf("[WARNING] Failed hashing in encrypt: %s", err)
		return []byte{}, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("[WARNING] Error generating ciphertext: %s", err)
		return []byte{}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[WARNING] Error creating new GCM from block: %s", err)
		return []byte{}, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("[WARNING] Error reading GCM nonce: %s", err)
		return []byte{}, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// base64 encoding to ensure we can store it as a string
	parsedValue := base64.StdEncoding.EncodeToString(ciphertext)
	return []byte(parsedValue), nil
}

func HandleKeyDecryption(data []byte, passphrase string) ([]byte, error) {
	key, err := create32Hash(passphrase)
	if err != nil {
		log.Printf("[ERROR] Failed hashing in decrypt: %s", err)
		return []byte{}, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("[ERROR] Error creating cipher from key in decryption: %s", err)
		return []byte{}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[ERROR] Error creating new GCM block in decryption: %s", err)
		return []byte{}, err
	}

	parsedData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		//log.Printf("[WARNING] Failed base64 decode for auth key '%s': '%s'. Data: '%s'. Returning as if this is valid.", data, err, string(data))
		//return []byte{}, err
		return data, nil
	}

	nonceSize := gcm.NonceSize()
	if nonceSize > len(parsedData) {
		log.Printf("[ERROR] Nonce size is larger than parsed data. Returning as if this is valid.")
		return data, nil
	}

	nonce, ciphertext := parsedData[:nonceSize], parsedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Printf("[ERROR] Error reading decryptionkey: %s", err)
		return []byte{}, err
	}

	return plaintext, nil
}

func HandleListCacheKeys(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[DEBUG] Api authentication failed in list cache keys: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
		return
	}

	if user.Role != "admin" {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Admin required"}`))
		return
	}

	//for key, value := range data.Apps {
	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	ctx := GetContext(request)
	org, err := GetOrg(ctx, fileId)
	if err != nil {
		log.Printf("[INFO] Organization doesn't exist: %s", err)
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	maxAmount := 20
	top, topOk := request.URL.Query()["top"]
	if topOk && len(top) > 0 {
		val, err := strconv.Atoi(top[0])
		if err == nil {
			maxAmount = val
		}
	}

	cursor := ""
	cursorList, cursorOk := request.URL.Query()["cursor"]
	if cursorOk && len(cursorList) > 0 {
		cursor = cursorList[0]
	}

	keys, newCursor, err := GetAllCacheKeys(ctx, org.Id, maxAmount, cursor)
	isSuccess := true
	if err != nil {
		isSuccess = false
	}

	newReturn := CacheReturn{
		Success: isSuccess,
		Keys:    keys,
		Cursor:  newCursor,
	}

	b, err := json.Marshal(newReturn)
	if err != nil {
		log.Printf("[WARNING] Failed to marshal cache keys for org %s: %s", org.Id, err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Something went wrong in cache key json management. Please refresh."}`))
		return
	}

	if err != nil {
		log.Printf("[INFO] Failed getting keys for org %s: %s", org.Id, err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write(b)
}

func HandleDeleteCacheKey(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[DEBUG] Api authentication failed in delete cache key: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
		return
	}

	//for key, value := range data.Apps {
	var orgId string
	var cacheKey string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		orgId = location[4]
		cacheKey = location[6]
	}

	if len(cacheKey) == 0 || len(orgId) == 0 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Missing org id or cache key"}`))
		return
	}

	ctx := GetContext(request)
	if orgId != user.ActiveOrg.Id {
		log.Printf("[INFO] OrgId %s and %s don't match", orgId, user.ActiveOrg.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Organization ID's don't match"}`))
		return
	}

	cacheKey, err = url.QueryUnescape(strings.Trim(cacheKey, " "))
	if err != nil {
		log.Printf("[WARNING] Failed to unescape cache key %s: %s", err)
		cacheKey = strings.Trim(cacheKey, " ")
	}

	cacheKey = strings.Replace(cacheKey, "%20", " ", -1)
	cacheKey = strings.Trim(cacheKey, " ")
	cacheId := fmt.Sprintf("%s_%s", orgId, cacheKey)

	cacheData, err := GetCacheKey(ctx, cacheId)
	if err != nil || cacheData.Key == "" {
		log.Printf("[WARNING] Failed to GET cache key '%s' for org %s (delete)", cacheId, orgId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed to get key. Does it exist?"}`))
		return
	}

	if cacheData.OrgId != user.ActiveOrg.Id {
		log.Printf("[INFO] OrgId '%s' and '%s' don't match", cacheData.OrgId, user.ActiveOrg.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Organization ID's don't match"}`))
		return
	}

	entity := "org_cache"

	DeleteKey(ctx, entity, cacheId)
	if len(cacheData.WorkflowId) > 0 {
		escapedKey := url.QueryEscape(cacheKey)

		DeleteKey(ctx, entity, fmt.Sprintf("%s_%s_%s", orgId, cacheData.WorkflowId, cacheData.Key))
		DeleteKey(ctx, entity, fmt.Sprintf("%s_%s_%s", orgId, cacheData.WorkflowId, escapedKey))

		DeleteKey(ctx, entity, fmt.Sprintf("%s_%s", cacheData.WorkflowId, cacheData.Key))

		DeleteKey(ctx, entity, fmt.Sprintf("%s_%s", cacheData.WorkflowId, escapedKey))
	}

	DeleteCache(ctx, cacheKey)
	DeleteCache(ctx, fmt.Sprintf("%s_%s", entity, cacheKey))
	DeleteCache(ctx, fmt.Sprintf("%s_%s", entity, orgId))

	log.Printf("[INFO] Successfully Deleted key '%s' for org %s", cacheKey, orgId)
	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

func HandleGetCacheKey(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed reading body"}`))
		return
	}

	//for key, value := range data.Apps {
	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	var tmpData CacheKeyData
	err = json.Unmarshal(body, &tmpData)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshalling in GET value: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if tmpData.OrgId != fileId {
		log.Printf("[INFO] OrgId %s and %s don't match", tmpData.OrgId, fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Organization ID's don't match"}`))
		return
	}

	ctx := GetContext(request)

	org, err := GetOrg(ctx, tmpData.OrgId)
	if err != nil {
		log.Printf("[INFO] Organization doesn't exist: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	workflowExecution, err := GetWorkflowExecution(ctx, tmpData.ExecutionId)
	if err != nil {
		log.Printf("[INFO] Failed getting the execution: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "No permission to get execution"}`))
		return
	}

	// Allows for execution auth AND user auth
	if workflowExecution.Authorization != tmpData.Authorization {
		// Get the user?
		user, err := HandleApiAuthentication(resp, request)
		if err != nil {
			log.Printf("[INFO] Execution auth %s and %s don't match", workflowExecution.Authorization, tmpData.Authorization)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
			return
		} else {
			if user.ActiveOrg.Id != org.Id {
				log.Printf("[INFO] Execution auth %s and %s don't match (2)", workflowExecution.Authorization, tmpData.Authorization)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
				return
			}
		}
	}

	if workflowExecution.Status != "EXECUTING" {
		log.Printf("[INFO] Workflow %s isn't executing and shouldn't be searching", workflowExecution.ExecutionId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow isn't executing"}`))
		return
	}

	if workflowExecution.ExecutionOrg != org.Id {
		log.Printf("[INFO] Org %s wasn't used to execute %s", org.Id, workflowExecution.ExecutionId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Bad organization specified"}`))
		return
	}

	tmpData.Key = strings.Trim(tmpData.Key, " ")
	cacheId := fmt.Sprintf("%s_%s", tmpData.OrgId, tmpData.Key)
	cacheData, err := GetCacheKey(ctx, cacheId)
	if err != nil {
		log.Printf("[WARNING] Failed to GET cache key %s for org %s (get)", tmpData.Key, tmpData.OrgId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed to get key. Does it exist?"}`))
		return
	}

	cacheData.Success = true
	cacheData.ExecutionId = ""
	cacheData.Authorization = ""
	cacheData.OrgId = ""

	log.Printf("[INFO] Successfully GOT key '%s' for org %s", tmpData.Key, tmpData.OrgId)
	b, err := json.Marshal(cacheData)
	if err != nil {
		log.Printf("[WARNING] Failed to GET cache key %s for org %s", tmpData.Key, tmpData.OrgId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed to get key. Does it exist?"}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write(b)
}

func HandleSetCacheKey(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, usererr := HandleApiAuthentication(resp, request)

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Failed reading body in set cache: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed reading body"}`))
		return
	}

	//for key, value := range data.Apps {
	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	var tmpData CacheKeyData
	err = json.Unmarshal(body, &tmpData)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshalling in setvalue: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)

	if len(tmpData.OrgId) == 0 {
		//log.Printf("[INFO] No org id specified. User org: %#v", user.ActiveOrg)
		tmpData.OrgId = user.ActiveOrg.Id
	}

	org, err := GetOrg(ctx, tmpData.OrgId)
	if err != nil {
		log.Printf("[WARNING] Organization doesn't exist: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	workflowExecution, err := GetWorkflowExecution(ctx, tmpData.ExecutionId)
	if err != nil {
		if len(tmpData.ExecutionId) > 0 {
			log.Printf("[WARNING] Failed getting exec: %s", err)
			resp.WriteHeader(500)
			resp.Write([]byte(`{"success": false, "reason": "No permission to get execution"}`))
			return
		}

		workflowExecution.Authorization = uuid.NewV4().String()
	}

	// Allows for execution auth AND user auth
	if workflowExecution.Authorization != tmpData.Authorization {

		// Get the user?
		if usererr != nil {
			log.Printf("[INFO] Execution auth %s and %s don't match", workflowExecution.Authorization, tmpData.Authorization)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
			return
		} else {
			if user.ActiveOrg.Id != org.Id {
				log.Printf("[INFO] Execution auth %s and %s don't match (2)", workflowExecution.Authorization, tmpData.Authorization)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
				return
			}

			tmpData.OrgId = user.ActiveOrg.Id
		}
	} else {
		if workflowExecution.Status != "EXECUTING" {
			log.Printf("[INFO] Workflow '%s' isn't executing and shouldn't be searching", workflowExecution.ExecutionId)
			resp.WriteHeader(400)
			resp.Write([]byte(`{"success": false, "reason": "Workflow isn't executing"}`))
			return
		}

		if workflowExecution.ExecutionOrg != org.Id {
			log.Printf("[INFO] Org '%s' wasn't used to execute %s", org.Id, workflowExecution.ExecutionId)
			resp.WriteHeader(403)
			resp.Write([]byte(`{"success": false, "reason": "Bad organization specified"}`))
			return
		}
	}

	if tmpData.OrgId != fileId {
		log.Printf("[INFO] OrgId '%s' and '%s' don't match (set cache)", tmpData.OrgId, fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Organization ID's don't match"}`))
		return
	}

	if len(tmpData.Value) == 0 {
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false, "reason": "Value can't be empty"}`))
		return
	}

	tmpData.Key = strings.Trim(tmpData.Key, " ")
	err = SetCacheKey(ctx, tmpData)
	if err != nil {
		log.Printf("[WARNING] Failed to set cache key '%s' for org %s", tmpData.Key, tmpData.OrgId)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "Failed to set data"}`))
		return
	}

	log.Printf("[INFO] Successfully set key '%s' for org '%s' (%s)", tmpData.Key, org.Name, tmpData.OrgId)
	type returnStruct struct {
		Success bool `json:"success"`
	}

	returnData := returnStruct{
		Success: true,
	}

	b, err := json.Marshal(returnData)
	if err != nil {
		b = []byte(`{"success": true}`)
	}

	resp.WriteHeader(200)
	resp.Write(b)
}

// Checks authentication string for Webhooks
func CheckHookAuth(request *http.Request, auth string) error {
	if len(auth) == 0 {
		return nil
	}

	log.Printf("[INFO] Checking hook auth: %s", auth)
	// Print headers
	for name, headers := range request.Header {
		name = strings.ToLower(name)
		log.Printf("[INFO] %s = %s", name, headers)
	}

	authSplit := strings.Split(auth, "\n")
	for _, line := range authSplit {
		lineSplit := strings.Split(line, "=")
		if strings.Contains(line, ":") {
			lineSplit = strings.Split(line, ":")
		}

		if len(lineSplit) >= 2 {
			validationHeader := strings.ToLower(strings.TrimSpace(lineSplit[0]))
			found := false

			joinedItemValue := strings.Join(lineSplit[1:], "=")
			log.Printf("[INFO] Checking %s = %s", validationHeader, joinedItemValue)
			for key, value := range request.Header {
				if strings.ToLower(key) == validationHeader && len(value) > 0 {
					//log.Printf("FOUND KEY %s. Value: %s", validationHeader, value)
					if value[0] == strings.TrimSpace(joinedItemValue) {
						found = true
						break
					}
				}
			}

			if !found {
				return errors.New(fmt.Sprintf("Missing or bad header: %s", validationHeader))
			}

			//log.Printf("Find header %s", validationHeader)
			//itemHeader := request.Header[validationHeader]
			//log.Printf("LINE: %s. Header: %s", line, itemHeader)
		} else {
			log.Printf("[WARNING] Bad auth line: %s. NOT checking auth.", line)
		}
	}

	//return errors.New("Bad auth!")
	return nil
}

// Fileid = the app to execute
// Body = The action body received from the user to test.
func PrepareSingleAction(ctx context.Context, user User, fileId string, body []byte) (WorkflowExecution, error) {
	var action Action

	workflowExecution := WorkflowExecution{}
	err := json.Unmarshal(body, &action)
	if err != nil {
		log.Printf("[WARNING] Failed action single execution unmarshaling: %s", err)
		return workflowExecution, err
	}

	/*
		if len(workflow.Name) > 0 || len(workflow.Owner) > 0 || len(workflow.OrgId) > 0 || len(workflow.Actions) != 1 {
			log.Printf("[WARNING] Bad length for some characteristics in single execution of %s", fileId)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	*/

	if fileId != action.AppID {
		log.Printf("[WARNING] Bad appid in single execution of App %s", fileId)
		return workflowExecution, err
	}

	if len(action.ID) == 0 {
		action.ID = uuid.NewV4().String()
	}

	app, err := GetApp(ctx, fileId, user, false)
	if err != nil {
		log.Printf("[WARNING] Error getting app (execute SINGLE app action): %s", fileId)
		return workflowExecution, err
	}

	if app.Authentication.Required && len(action.AuthenticationId) == 0 {
		log.Printf("[WARNING] Tried to execute SINGLE %s WITHOUT auth (missing)", app.Name)

		found := false
		for _, param := range action.Parameters {
			if param.Configuration {
				found = true
				break
			}
		}

		if !found {
			return workflowExecution, errors.New("You must authenticate the app first")
		}
	}

	newParams := []WorkflowAppActionParameter{}
	for _, param := range action.Parameters {
		if param.Configuration && len(param.Value) == 0 {
			continue
		}

		newParams = append(newParams, param)
	}

	isOauth2 := false
	if len(action.AuthenticationId) > 0 {
		curAuth, err := GetWorkflowAppAuthDatastore(ctx, action.AuthenticationId)
		if err != nil {
			log.Printf("[ERROR] Failed getting authentication with ID %s for org %s: %s", action.AuthenticationId, user.ActiveOrg.Id, err)
			return workflowExecution, err
		}

		if curAuth.Type == "oauth2" {
			isOauth2 = true

			//for _, auth := range curAuth.Fields {
			//	log.Printf("Oauth2 Field: %s", auth.Key)
			//}
		}

		if user.ActiveOrg.Id != curAuth.OrgId {
			log.Printf("[WARNING] User '%s' (%s) in org %s tried to use bad auth %s from org %s during execution", user.Username, user.Id, user.ActiveOrg.Id, action.AuthenticationId, curAuth.OrgId)
			return workflowExecution, err
		}

		// Handle fixing of order for the fields

		/*
			lastAccesstoken := -1
			lastRefreshtoken := -1
			for cnt, param := range action.Parameters {
				if param.Type == "access_token" {
					lastAccesstoken = cnt
				}

				if param.Type == "refresh_token" {
					lastRefreshtoken = cnt
				}
			}
		*/

		newFields := []AuthenticationStore{}
		if curAuth.Encrypted {
			for _, field := range curAuth.Fields {
				parsedKey := fmt.Sprintf("%s_%d_%s_%s", curAuth.OrgId, curAuth.Created, curAuth.Label, field.Key)
				newValue, err := HandleKeyDecryption([]byte(field.Value), parsedKey)
				if err != nil {
					log.Printf("[ERROR] Failed decryption for %s: %s", field.Key, err)
					break
				}

				if field.Key == "url" {
					//log.Printf("Value2 (%s): %s", field.Key, string(newValue))
					if strings.HasSuffix(string(newValue), "/") {
						newValue = []byte(string(newValue)[0 : len(newValue)-1])
					}
				}

				newParam := WorkflowAppActionParameter{
					Name:  field.Key,
					ID:    action.AuthenticationId,
					Value: string(newValue),
				}

				field.Value = string(newValue)
				newFields = append(newFields, field)
				newParams = append(newParams, newParam)
			}

		} else {
			newFields = curAuth.Fields
			//log.Printf("[INFO] AUTH IS NOT ENCRYPTED - attempting auto-encrypting if key is set!")
			//err = SetWorkflowAppAuthDatastore(ctx, curAuth, curAuth.Id)
			//if err != nil {
			//	log.Printf("[WARNING] Failed running encryption during execution: %s", err)
			//}
			for _, auth := range curAuth.Fields {

				newParam := WorkflowAppActionParameter{
					Name:  auth.Key,
					ID:    action.AuthenticationId,
					Value: auth.Value,
				}

				newParams = append(newParams, newParam)
			}
		}

		if isOauth2 {
			log.Printf("\n[INFO] OAUTH2: %s\n", curAuth.Type)
			curAuth.Fields = newFields

			newAuth, err := RunOauth2Request(ctx, user, *curAuth, true)
			if err != nil {
				log.Printf("[ERROR] Failed running single action oauth2 refresh request for org %s: %s", user.ActiveOrg.Id, err)
			} else {
				curAuth = &newAuth
			}

			// Fix params from here? Did access token get added properly?
		}

		// Rebuild params with the right data. This is to prevent issues on the frontend

		action.Parameters = newParams
	}

	for _, param := range action.Parameters {
		newName := GetValidParameters([]string{param.Name})
		if len(newName) > 0 {
			param.Name = newName[0]
		}

		if param.Required && len(param.Value) == 0 {
			if param.Name == "username_basic" {
				param.Name = "username"
			} else if param.Name == "password_basic" {
				param.Name = "password"
			}

			param.Name = strings.Replace(param.Name, "_", " ", -1)
			param.Name = strings.Title(param.Name)

			value := fmt.Sprintf("Param %s can't be empty. Please fill all required parameters (orange outline). If you don't know the value, input space in the field.", param.Name)
			log.Printf("[WARNING] During single exec: %s", value)
			return workflowExecution, errors.New(value)
		}

		newParams = append(newParams, param)
	}

	action.Parameters = newParams

	if isOauth2 {

		foundAction := WorkflowAppAction{}
		for _, inner := range app.Actions {
			//log.Printf("[INFO] Checking action %s vs %s", inner.Name, action.Name)
			if inner.Name == action.Name {
				foundAction = inner
				break
			}
		}

		if len(foundAction.Name) > 0 {
			// Find the original and see if this matches?
			// Seems wrong from Oauth2 :thinking:
			newParams := []WorkflowAppActionParameter{}
			for _, param := range action.Parameters {
				//log.Printf("[DEBUG] Param: %s", param.Name)
				found := false
				for _, foundActionParam := range foundAction.Parameters {
					if foundActionParam.Name == param.Name {
						found = true
						newParams = append(newParams, param)
						break
					}
				}

				if !found {
					//log.Printf("[WARNING] Auth Param %s not found in action %s", param.Name, foundAction.Name)
					if param.Name == "access_token" {
						newParams = append(newParams, param)
					}
				}
			}

			action.Parameters = newParams
		}

		/*
			for _, param := range action.Parameters {
				log.Printf("[DEBUG] Param2: %s", param.Name)
			}
		*/
	}

	action.Sharing = app.Sharing
	action.Public = app.Public
	action.Generated = app.Generated

	workflow := Workflow{
		Actions: []Action{
			action,
		},
		Start:     action.ID,
		ID:        uuid.NewV4().String(),
		Generated: true,
		Hidden:    true,
	}

	//log.Printf("Sharing: %s, Public: %s, Generated: %s. Start: %s", action.Sharing, action.Public, action.Generated, workflow.Start)

	workflowExecution = WorkflowExecution{
		Workflow:      workflow,
		Start:         workflow.Start,
		ExecutionId:   uuid.NewV4().String(),
		WorkflowId:    workflow.ID,
		StartedAt:     int64(time.Now().Unix()),
		CompletedAt:   0,
		Authorization: uuid.NewV4().String(),
		Status:        "EXECUTING",
	}

	if user.ActiveOrg.Id != "" {
		workflow.ExecutingOrg = user.ActiveOrg
		workflowExecution.ExecutionOrg = user.ActiveOrg.Id
		workflowExecution.OrgId = user.ActiveOrg.Id
	}

	err = SetWorkflowExecution(ctx, workflowExecution, true)
	if err != nil {
		log.Printf("[WARNING] Failed handling single execution setup: %s", err)
		return workflowExecution, err
	}

	return workflowExecution, nil
}

func HandleRetValidation(ctx context.Context, workflowExecution WorkflowExecution, resultAmount int) SingleResult {
	cnt := 0

	returnBody := SingleResult{
		Success:       true,
		Id:            workflowExecution.ExecutionId,
		Authorization: workflowExecution.Authorization,
		Result:        "",
		Errors:        []string{},
	}

	// VERY short sleeptime here on purpose
	maxSeconds := 10
	sleeptime := 25
	for {
		time.Sleep(25 * time.Millisecond)
		newExecution, err := GetWorkflowExecution(ctx, workflowExecution.ExecutionId)
		if err != nil {
			log.Printf("[WARNING] Failed getting single execution data: %s", err)
			break
		}

		if len(newExecution.Results) > resultAmount-1 {
			relevantIndex := len(newExecution.Results) - 1
			if len(newExecution.Results[relevantIndex].Result) > 0 {
				returnBody.Result = newExecution.Results[relevantIndex].Result

				// Check for single action errors in liquid and similar
				// to be used in the frontend
				for _, param := range newExecution.Results[relevantIndex].Action.Parameters {
					//log.Printf("Name: %s", param.Name)
					if strings.Contains(param.Name, "liquid") && !ArrayContains(returnBody.Errors, param.Value) {
						returnBody.Errors = append(returnBody.Errors, param.Value)
					}
				}
				break
			}
		}

		cnt += 1
		//log.Printf("Cnt: %d", cnt)
		if cnt == (maxSeconds * (maxSeconds * 100 / sleeptime)) {
			break
		}
	}

	if len(returnBody.Result) == 0 {
		returnBody.Success = false
	}

	return returnBody

}

func GetDocs(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	location := strings.Split(request.URL.String(), "/")
	if len(location) < 5 {
		resp.WriteHeader(404)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Bad path. Use e.g. /api/v1/docs/workflows.md"`)))
		return
	}

	if strings.Contains(location[4], "?") {
		location[4] = strings.Split(location[4], "?")[0]
	}

	ctx := GetContext(request)
	downloadLocation, downloadOk := request.URL.Query()["location"]
	version, versionOk := request.URL.Query()["version"]
	cacheKey := fmt.Sprintf("docs_%s", location[4])
	if downloadOk {
		cacheKey = fmt.Sprintf("%s_%s", cacheKey, downloadLocation[0])
	}

	if versionOk {
		cacheKey = fmt.Sprintf("%s_%s", cacheKey, version[0])
	}

	cache, err := GetCache(ctx, cacheKey)
	if err == nil {
		cacheData := []byte(cache.([]uint8))
		resp.WriteHeader(200)
		resp.Write(cacheData)
		return
	}

	owner := "shuffle"
	repo := "shuffle-docs"
	path := "docs"
	docPath := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/%s/%s.md", owner, repo, path, location[4])

	// FIXME: User controlled and dangerous (possibly). Uses Markdown on the frontend to render it
	realPath := ""
	//log.Printf("\n\n INSIDe Download path (%s): %s with version %s!\n\n", location[4], downloadLocation, version)

	newname := location[4]
	if downloadOk {
		if downloadLocation[0] == "openapi" {
			newname = strings.ReplaceAll(strings.ToLower(location[4]), `%20`, "_")
			docPath = fmt.Sprintf("https://raw.githubusercontent.com/Shuffle/openapi-apps/master/docs/%s.md", newname)
			realPath = fmt.Sprintf("https://github.com/Shuffle/openapi-apps/blob/master/docs/%s.md", newname)

		} else if downloadLocation[0] == "python" && versionOk {
			// Apparently this uses dashes for no good reason?
			// Should maybe move everything over to underscores later?
			newname = strings.ReplaceAll(newname, `%20`, "-")
			newname = strings.ReplaceAll(newname, ` `, "-")
			newname = strings.ReplaceAll(newname, `_`, "-")
			newname = strings.ToLower(newname)

			if version[0] == "1.0.0" {
				docPath = fmt.Sprintf("https://raw.githubusercontent.com/Shuffle/python-apps/master/%s/1.0.0/README.md", newname)
				realPath = fmt.Sprintf("https://github.com/Shuffle/python-apps/blob/master/%s/1.0.0/README.md", newname)

				log.Printf("[INFO] Should download python app for version %s: %s", version[0], docPath)

			} else {
				realPath = fmt.Sprintf("https://github.com/Shuffle/python-apps/blob/master/%s/README.md", newname)
				docPath = fmt.Sprintf("https://raw.githubusercontent.com/Shuffle/python-apps/master/%s/README.md", newname)
			}

		}
	}

	//log.Printf("Docpath: %s", docPath)

	httpClient := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		docPath,
		nil,
	)

	if err != nil {
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Bad path. Use e.g. /api/v1/docs/workflows.md"}`)))
		resp.WriteHeader(404)
		return
	}

	newresp, err := httpClient.Do(req)
	if err != nil {
		resp.WriteHeader(404)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Bad path. Use e.g. /api/v1/docs/workflows.md"}`)))
		return
	}

	body, err := ioutil.ReadAll(newresp.Body)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can't parse data"}`)))
		return
	}

	commitOptions := &github.CommitsListOptions{
		Path: fmt.Sprintf("%s/%s.md", path, location[4]),
	}

	parsedLink := fmt.Sprintf("https://github.com/%s/%s/blob/master/%s/%s.md", owner, repo, path, location[4])
	if len(realPath) > 0 {
		parsedLink = realPath
	}

	client := github.NewClient(nil)
	githubResp := GithubResp{
		Name:         location[4],
		Contributors: []GithubAuthor{},
		Edited:       "",
		ReadTime:     len(body) / 10 / 250,
		Link:         parsedLink,
	}

	if githubResp.ReadTime == 0 {
		githubResp.ReadTime = 1
	}

	info, _, err := client.Repositories.ListCommits(ctx, owner, repo, commitOptions)
	if err != nil {
		log.Printf("[WARNING] Failed getting commit info: %s", err)
	} else {
		//log.Printf("Info: %s", info)
		for _, commit := range info {
			//log.Printf("Commit: %s", commit.Author)
			newAuthor := GithubAuthor{}
			if commit.Author != nil && commit.Author.AvatarURL != nil {
				newAuthor.ImageUrl = *commit.Author.AvatarURL
			}

			if commit.Author != nil && commit.Author.HTMLURL != nil {
				newAuthor.Url = *commit.Author.HTMLURL
			}

			found := false
			for _, contributor := range githubResp.Contributors {
				if contributor.Url == newAuthor.Url {
					found = true
					break
				}
			}

			if !found && len(newAuthor.Url) > 0 && len(newAuthor.ImageUrl) > 0 {
				githubResp.Contributors = append(githubResp.Contributors, newAuthor)
			}
		}
	}

	type Result struct {
		Success bool       `json:"success"`
		Reason  string     `json:"reason"`
		Meta    GithubResp `json:"meta"`
	}

	var result Result
	result.Success = true
	result.Meta = githubResp

	//applog.Infof(ctx, string(body))
	//applog.Infof(ctx, "Url: %s", docPath)
	//log.Printf("[INFO] GOT BODY OF LENGTH %d", len(string(body)))

	result.Reason = string(body)
	b, err := json.Marshal(result)
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}

	// Not caching 404s
	//if Result.Success && !strings.Contains(string(b), "404: Not Found") {
	err = SetCache(ctx, cacheKey, b, 30)
	if err != nil {
		log.Printf("[WARNING] Failed setting cache for doc %s: %s", location[4], err)
	}

	resp.WriteHeader(200)
	resp.Write(b)
}

func GetDocList(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	ctx := GetContext(request)
	cacheKey := "docs_list"
	cache, err := GetCache(ctx, cacheKey)
	result := FileList{}
	if err == nil {
		cacheData := []byte(cache.([]uint8))
		resp.WriteHeader(200)
		resp.Write(cacheData)
		return
	}

	client := github.NewClient(nil)
	owner := "shuffle"
	repo := "shuffle-docs"
	path := "docs"
	_, item1, _, err := client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Error listing directory"}`)))
		return
	}

	if len(item1) == 0 {
		resp.WriteHeader(500)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "No docs available."}`)))
		return
	}

	names := []GithubResp{}
	for _, item := range item1 {
		if !strings.HasSuffix(*item.Name, "md") {
			continue
		}

		// Average word length = 5. Space = 1. 5+1 = 6 avg.
		// Words = *item.Size/6/250
		//250 = average read time / minute
		// Doubling this for bloat removal in Markdown~
		// Should fix this lol
		githubResp := GithubResp{
			Name:         (*item.Name)[0 : len(*item.Name)-3],
			Contributors: []GithubAuthor{},
			Edited:       "",
			ReadTime:     *item.Size / 6 / 250,
			Link:         fmt.Sprintf("https://github.com/%s/%s/blob/master/%s/%s", owner, repo, path, *item.Name),
		}

		names = append(names, githubResp)
	}

	//log.Printf(names)
	result.Success = true
	result.Reason = "Success"
	result.List = names
	b, err := json.Marshal(result)
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}

	err = SetCache(ctx, cacheKey, b, 30)
	if err != nil {
		log.Printf("[WARNING] Failed setting cache for cachekey %s: %s", cacheKey, err)
	}

	resp.WriteHeader(200)
	resp.Write(b)
}

func md5sum(data []byte) string {
	hasher := md5.New()
	hasher.Write(data)
	newmd5 := hex.EncodeToString(hasher.Sum(nil))
	return newmd5
}

// Checks if data is sent from Worker >0.8.51, which sends a full execution
// instead of individial results
func ValidateNewWorkerExecution(ctx context.Context, body []byte) error {
	var execution WorkflowExecution
	err := json.Unmarshal(body, &execution)
	if err != nil {
		log.Printf("[WARNING] Failed execution unmarshaling: %s", err)
		if strings.Contains(fmt.Sprintf("%s", err), "array into") {
			log.Printf("Array unmarshal error: %s", string(body))
		}

		return err
	}
	//log.Printf("\n\nGOT EXEC WITH RESULT %s (%d)\n\n", execution.Status, len(execution.Results))

	baseExecution, err := GetWorkflowExecution(ctx, execution.ExecutionId)
	if err != nil {
		log.Printf("[ERROR] Failed getting execution (workflowqueue) %s: %s", execution.ExecutionId, err)
		return err
	}

	if baseExecution.Authorization != execution.Authorization {
		return errors.New("Bad authorization when validating execution")
	}

	// used to validate if it's actually the right marshal
	if len(baseExecution.Workflow.Actions) != len(execution.Workflow.Actions) {
		return errors.New(fmt.Sprintf("Bad length of actions (probably normal app): %d", len(execution.Workflow.Actions)))
	}

	if len(baseExecution.Workflow.Triggers) != len(execution.Workflow.Triggers) {
		return errors.New(fmt.Sprintf("Bad length of trigger: %d (probably normal app)", len(execution.Workflow.Triggers)))
	}

	//if len(baseExecution.Results) >= len(execution.Results) {
	if len(baseExecution.Results) > len(execution.Results) {
		return errors.New(fmt.Sprintf("Can't have less actions in a full execution than what exists: %d (old) vs %d (new)", len(baseExecution.Results), len(execution.Results)))
	}

	//if baseExecution.Status != "WAITING" && baseExecution.Status != "EXECUTING" {
	//	return errors.New(fmt.Sprintf("Workflow is already finished or failed. Can't update"))
	//}

	if execution.Status == "EXECUTING" {
		//log.Printf("[INFO] Inside executing.")
		extra := 0
		for _, trigger := range execution.Workflow.Triggers {
			//log.Printf("Appname trigger (0): %s", trigger.AppName)
			if trigger.AppName == "User Input" || trigger.AppName == "Shuffle Workflow" {
				extra += 1
			}
		}

		if len(execution.Workflow.Actions)+extra == len(execution.Results) {
			execution.Status = "FINISHED"
		}
	}

	// Finds if subflow HAS a value when it should, otherwise it's not being set
	//log.Printf("\n\nUpdating worker execution info")
	for _, result := range execution.Results {
		//log.Printf("%s = %s", result.Action.AppName, result.Status)
		if result.Action.AppName == "shuffle-subflow" {
			if result.Status == "SKIPPED" {
				continue
			}

			//log.Printf("\n\nFound SUBFLOW in full result send \n\n")
			for _, trigger := range baseExecution.Workflow.Triggers {
				if trigger.ID == result.Action.ID {
					//log.Printf("Found SUBFLOW id: %s", trigger.ID)

					for _, param := range trigger.Parameters {
						if param.Name == "check_result" && param.Value == "true" {
							//log.Printf("Found check as true!")

							var subflowData SubflowData
							err = json.Unmarshal([]byte(result.Result), &subflowData)
							if err != nil {
								log.Printf("Failed unmarshal in subflow check for %s: %s", result.Result, err)
							} else if len(subflowData.Result) == 0 {
								log.Printf("There is no result yet. Don't save?")
							} else {
								//log.Printf("There is a result: %s", result.Result)
							}

							break
						}
					}

					break
				}
			}
		}
	}

	// Check status is finished, and set timestamp for finished if it's 0
	if execution.Status == "FINISHED" || execution.Status == "ABORTED" || execution.Status == "FAILED" {
		if baseExecution.CompletedAt == 0 {
			baseExecution.CompletedAt = time.Now().Unix()
		}
	}

	// FIXME: Add extra here
	//executionLength := len(baseExecution.Workflow.Actions)
	//if executionLength != len(execution.Results) {
	//	return errors.New(fmt.Sprintf("Bad length of actions vs results: want: %d have: %d", executionLength, len(execution.Results)))
	//}

	err = SetWorkflowExecution(ctx, execution, true)
	executionSet := true
	if err == nil {
		log.Printf("[INFO][%s] Set workflowexecution based on new worker (>0.8.53) for workflow %s. Actions: %d, Triggers: %d, Results: %d, Status: %s", execution.ExecutionId, execution.WorkflowId, len(execution.Workflow.Actions), len(execution.Workflow.Triggers), len(execution.Results), execution.Status) //, execution.Result)
		executionSet = true
	} else {
		log.Printf("[WARNING] Failed setting the execution for new worker (>0.8.53) - retrying once: %s. ExecutionId: %s, Actions: %d, Triggers: %d, Results: %d, Status: %s", err, execution.ExecutionId, len(execution.Workflow.Actions), len(execution.Workflow.Triggers), len(execution.Results), execution.Status)
		// Retrying
		time.Sleep(5 * time.Second)
		err = SetWorkflowExecution(ctx, execution, true)
		if err != nil {
			log.Printf("[ERROR] Failed setting the execution for new worker (>0.8.53) - 2nd attempt: %s. ExecutionId: %s, Actions: %d, Triggers: %d, Results: %d, Status: %s", err, execution.ExecutionId, len(execution.Workflow.Actions), len(execution.Workflow.Triggers), len(execution.Results), execution.Status)
		} else {
			executionSet = true
		}
	}

	// Long convoluted way of validating and setting the value of a subflow that is also a loop
	// FIXME: May cause errors in worker that runs it all instantly due to
	// timing issues / non-queues
	if executionSet {
		RunFixParentWorkflowResult(ctx, execution)
	}

	DeleteCache(ctx, fmt.Sprintf("workflowexecution_%s", execution.WorkflowId))
	DeleteCache(ctx, fmt.Sprintf("workflowexecution_%s_50", execution.WorkflowId))
	DeleteCache(ctx, fmt.Sprintf("workflowexecution_%s_100", execution.WorkflowId))

	return nil
}

func RunFixParentWorkflowResult(ctx context.Context, execution WorkflowExecution) error {
	//log.Printf("IS IT SUBFLOW?")
	if len(execution.ExecutionParent) > 0 && execution.Status != "EXECUTING" && (project.Environment == "onprem" || project.Environment == "cloud") {
		log.Printf("[DEBUG] Got the result '%s' for subflow of %s. Workflow: '%s' Check if this should be added to loop.", execution.Result, execution.ExecutionParent, execution.WorkflowId)

		parentExecution, err := GetWorkflowExecution(ctx, execution.ExecutionParent)
		if err == nil {
			isLooping := false
			setExecution := true
			shouldSetValue := false
			for _, trigger := range parentExecution.Workflow.Triggers {
				if trigger.ID == execution.ExecutionSourceNode {
					for _, param := range trigger.Parameters {
						if param.Name == "workflow" && param.Value != execution.Workflow.ID {
							setExecution = false
						}

						if param.Name == "argument" && strings.Contains(param.Value, "$") && strings.Contains(param.Value, ".#") {
							isLooping = true
						}

						if param.Name == "check_result" && param.Value == "true" {
							shouldSetValue = true
						}
					}

					break
				}
			}

			if !isLooping && setExecution && shouldSetValue && parentExecution.Status == "EXECUTING" {
				//log.Printf("[DEBUG] Its NOT looping. Should set?")
				return nil
			} else if isLooping && setExecution && shouldSetValue && parentExecution.Status == "EXECUTING" {
				log.Printf("[DEBUG] Parentexecutions' subflow IS looping and is correct workflow. Should find correct answer in the node's result. Length of results: %d", len(parentExecution.Results))
				// 1. Find the action's existing result
				// 2. ONLY update it if the action status is WAITING and workflow status is EXECUTING
				// 3. IF all parts of the subflow execution are finished, set it to FINISHED
				// 4. If result length == length of actions + extra, set it to FINISHED
				// 5. Before setting parent execution, make sure to grab the latest version of the workflow again, in case processing time is slow
				resultIndex := -1
				updateIndex := -1
				for parentResultIndex, result := range parentExecution.Results {
					if result.Action.ID != execution.ExecutionSourceNode {
						continue
					}
					log.Printf("[DEBUG] Found action %s' results: %s", result.Action.ID, result.Result)
					if result.Status != "WAITING" {
						break
					}

					//result.Result
					var subflowDataLoop []SubflowData
					err = json.Unmarshal([]byte(result.Result), &subflowDataLoop)
					if err != nil {
						log.Printf("[DEBUG] Failed unmarshaling in set parent data: %s", err)
						break
					}

					for subflowIndex, subflowResult := range subflowDataLoop {
						if subflowResult.ExecutionId != execution.ExecutionId {
							continue
						}

						log.Printf("[DEBUG] Found right execution on index %d. Result: %s", subflowIndex, subflowResult.Result)
						if len(subflowResult.Result) == 0 {
							updateIndex = subflowIndex
						}

						resultIndex = parentResultIndex
						break
					}
				}

				// FIXME: MAY cause transaction issues.
				if updateIndex >= 0 && resultIndex >= 0 {
					log.Printf("[DEBUG] Should update index %d in resultIndex %d with new result %s", updateIndex, resultIndex, execution.Result)
					// FIXME: Are results ordered? Hmmmmm
					// Again, get the result, just in case, and update that exact value instantly
					newParentExecution, err := GetWorkflowExecution(ctx, execution.ExecutionParent)
					if err == nil {

						var subflowDataLoop []SubflowData
						err = json.Unmarshal([]byte(newParentExecution.Results[resultIndex].Result), &subflowDataLoop)
						if err == nil {
							subflowDataLoop[updateIndex].Result = execution.Result
							subflowDataLoop[updateIndex].ResultSet = true

							marshalledSubflow, err := json.Marshal(subflowDataLoop)
							if err == nil {
								newParentExecution.Results[resultIndex].Result = string(marshalledSubflow)
								err = SetWorkflowExecution(ctx, *newParentExecution, true)
								if err != nil {
									log.Printf("[WARNING] Error saving parent execution in subflow setting: %s", err)
								} else {
									log.Printf("[DEBUG] Updated index %d in subflow result %d with value of length %d. IDS HAVE TO MATCH: %s vs %s", updateIndex, resultIndex, len(execution.Result), subflowDataLoop[updateIndex].ExecutionId, execution.ExecutionId)
								}
							}

							// Validating if all are done and setting back to executing
							allFinished := true
							for _, parentResult := range newParentExecution.Results {
								if parentResult.Action.ID != execution.ExecutionSourceNode {
									continue
								}

								var subflowDataLoop []SubflowData
								err = json.Unmarshal([]byte(parentResult.Result), &subflowDataLoop)
								if err == nil {
									for _, subflowResult := range subflowDataLoop {
										if subflowResult.ResultSet != true {
											allFinished = false
											break
										}
									}

									break
								} else {
									allFinished = false
									break
								}
							}

							// FIXME: This will break if subflow with loop is last node in two workflows in a row (main workflow -> []subflow -> []subflow)
							// Should it send the whole thing back as a result to itself to be handled manually? :thinking:
							if allFinished {
								newParentExecution.Results[resultIndex].Status = "SUCCESS"

								extra := 0
								for _, trigger := range newParentExecution.Workflow.Triggers {
									//log.Printf("Appname trigger (0): %s", trigger.AppName)
									if trigger.AppName == "User Input" || trigger.AppName == "Shuffle Workflow" {
										extra += 1
									}
								}

								if len(newParentExecution.Workflow.Actions)+extra == len(newParentExecution.Results) {
									newParentExecution.Status = "FINISHED"
								}

								err = SetWorkflowExecution(ctx, *newParentExecution, true)
								if err != nil {
									log.Printf("[ERROR] Failed updating setExecution to FINISHED and SUCCESS: %s", err)
								}
							}
						} else {
							log.Printf("[WARNING] Failed to unmarshal result in set parent subflow: %s", err)
						}

						//= newValue
					} else {
						log.Printf("[WARNING] Failed to update parent, because execution %s couldn't be found: %s", execution.ExecutionParent, err)
					}
				}
			}
		}
	}

	return nil
}

func fixCertificate(parsedX509Key string) string {
	parsedX509Key = strings.Replace(parsedX509Key, "&#13;", "", -1)
	if strings.Contains(parsedX509Key, "BEGIN CERT") && strings.Contains(parsedX509Key, "END CERT") {
		parsedX509Key = strings.Replace(parsedX509Key, "-----BEGIN CERTIFICATE-----\n", "", -1)
		parsedX509Key = strings.Replace(parsedX509Key, "-----BEGIN CERTIFICATE-----", "", -1)
		parsedX509Key = strings.Replace(parsedX509Key, "-----END CERTIFICATE-----\n", "", -1)
		parsedX509Key = strings.Replace(parsedX509Key, "-----END CERTIFICATE-----", "", -1)
	}

	// PingOne issue
	parsedX509Key = strings.Replace(parsedX509Key, "\r\n", "", -1)
	parsedX509Key = strings.Replace(parsedX509Key, "\n", "", -1)
	parsedX509Key = strings.Replace(parsedX509Key, "\r", "", -1)
	parsedX509Key = strings.Replace(parsedX509Key, " ", "", -1)
	parsedX509Key = strings.TrimSpace(parsedX509Key)
	//log.Printf("Len: %d", len(parsedX509Key))
	//log.Printf("%s", parsedX509Key)
	return parsedX509Key
}

// Example implementation of SSO, including a redirect for the user etc
// Should make this stuff only possible after login
func HandleOpenId(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	//https://dev-18062475.okta.com/oauth2/default/v1/authorize?client_id=oa3romteykJ2aMgx5d7&response_type=code&scope=openid&redirect_uri=http%3A%2F%2Flocalhost%3A5002%2Fapi%2Fv1%2Flogin_openid&state=state-296bc9a0-a2a2-4a57-be1a-d0e2fd9bb601&code_challenge_method=S256&code_challenge=codechallenge
	// http://localhost:5002/api/v1/login_openid?code=rrm8BS8eUIYpQWnoM_Lzh_QoT3-EwQ2c9YkjRcJWqk4&state=state-296bc9a0-a2a2-4a57-be1a-d0e2fd9bb601
	// http://localhost:5001/api/v1/login_openid#id_token=asdasd&session_state=asde9d78d8-6535-45fe-848d-0efa9f119595

	//code -> Token
	ctx := GetContext(request)

	skipValidation := false
	openidUser := OpenidUserinfo{}
	org := &Org{}
	code := request.URL.Query().Get("code")
	if len(code) == 0 {
		// Check id_token grant info
		if request.Method == "POST" {
			body, err := ioutil.ReadAll(request.Body)
			if err != nil {
				resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "No code or id_token specified - body read error in POST"}`)))
				resp.WriteHeader(401)
				return
			}

			stateSplit := strings.Split(string(body), "&")
			for _, innerstate := range stateSplit {
				itemsplit := strings.Split(innerstate, "=")

				if len(itemsplit) <= 1 {
					log.Printf("[WARNING] No key:value: %s", innerstate)
					continue
				}

				if itemsplit[0] == "id_token" {
					token, err := VerifyIdToken(ctx, itemsplit[1])
					if err != nil {
						log.Printf("[ERROR] Bad ID token provided: %s", err)
						resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Bad ID token provided"}`)))
						resp.WriteHeader(401)
						return
					}

					log.Printf("[DEBUG] Validated - token: %s!", token)
					openidUser.Sub = token.Sub
					org = &token.Org
					skipValidation = true

					break
				}
			}
		}

		if !skipValidation {
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "No code specified"}`)))
			resp.WriteHeader(401)
			return
		}
	}

	if !skipValidation {
		state := request.URL.Query().Get("state")
		if len(state) == 0 {
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "No state specified"}`)))
			return
		}

		stateBase, err := base64.StdEncoding.DecodeString(state)
		if err != nil {
			log.Printf("[ERROR] Failed base64 decode OpenID state: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed base64 decoding of state"}`)))
			return
		}

		log.Printf("State: %s", stateBase)
		foundOrg := ""
		foundRedir := ""
		foundChallenge := ""
		stateSplit := strings.Split(string(stateBase), "&")
		for _, innerstate := range stateSplit {
			itemsplit := strings.Split(innerstate, "=")
			//log.Printf("Itemsplit: %s", itemsplit)
			if len(itemsplit) <= 1 {
				log.Printf("[WARNING] No key:value: %s", innerstate)
				continue
			}

			if itemsplit[0] == "org" {
				foundOrg = strings.TrimSpace(itemsplit[1])
			}

			if itemsplit[0] == "redirect" {
				foundRedir = strings.TrimSpace(itemsplit[1])
			}

			if itemsplit[0] == "challenge" {
				foundChallenge = strings.TrimSpace(itemsplit[1])
			}
		}

		//log.Printf("Challenge len2: %d", len(foundChallenge))

		if len(foundOrg) == 0 {
			log.Printf("[ERROR] No org specified in state")
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "No org specified in state"}`)))
			return
		}

		org, err = GetOrg(ctx, foundOrg)
		if err != nil {
			log.Printf("[WARNING] Error getting org in OpenID: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Couldn't find the org for sign-in in Shuffle"}`))
			return
		}

		clientId := org.SSOConfig.OpenIdClientId
		tokenUrl := org.SSOConfig.OpenIdToken
		if len(tokenUrl) == 0 {
			log.Printf("[ERROR] No token URL specified for OpenID")
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "No token URL specified. Please make sure to specify a token URL in the /admin panel in Shuffle for OpenID Connect"}`)))
			return
		}

		//log.Printf("Challenge: %s", foundChallenge)
		body, err := RunOpenidLogin(ctx, clientId, tokenUrl, foundRedir, code, foundChallenge, org.SSOConfig.OpenIdClientSecret)
		if err != nil {
			log.Printf("[WARNING] Error with body read of OpenID Connect: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		openid := OpenidResp{}
		err = json.Unmarshal(body, &openid)
		if err != nil {
			log.Printf("[WARNING] Error in Openid marshal: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		// Automated replacement
		userInfoUrlSplit := strings.Split(org.SSOConfig.OpenIdAuthorization, "/")
		userinfoEndpoint := strings.Join(userInfoUrlSplit[0:len(userInfoUrlSplit)-1], "/") + "/userinfo"
		//userinfoEndpoint := strings.Replace(org.SSOConfig.OpenIdAuthorization, "/authorize", "/userinfo", -1)
		log.Printf("Userinfo endpoint: %s", userinfoEndpoint)
		client := &http.Client{}
		req, err := http.NewRequest(
			"GET",
			userinfoEndpoint,
			nil,
		)

		//req.Header.Add("accept", "application/json")
		//req.Header.Add("cache-control", "no-cache")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", openid.AccessToken))
		res, err := client.Do(req)
		if err != nil {
			log.Printf("[WARNING] OpenID client DO (2): %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed userinfo request"}`))
			return
		}

		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("[WARNING] OpenID client Body (2): %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed userinfo body parsing"}`))
			return
		}

		err = json.Unmarshal(body, &openidUser)
		if err != nil {
			log.Printf("[WARNING] Error in Openid marshal (2): %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	//log.Printf("Got user body: %s", string(body))

	/*

		BELOW HERE ITS ALL COPY PASTE OF USER INFO THINGS!

	*/

	if len(openidUser.Sub) == 0 {
		log.Printf("[WARNING] No user found in openid login (2)")
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if project.Environment == "cloud" {
		log.Printf("[WARNING] Openid SSO is not implemented for cloud yet. User %s", openidUser.Sub)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Cloud Openid is not available yet"}`)))
		return
	}

	userName := strings.ToLower(strings.TrimSpace(openidUser.Sub))
	if !strings.Contains(userName, "@") {
		log.Printf("[ERROR] Bad username, but allowing due to OpenID: %s. Full Subject: %#v", userName, openidUser)
	}
	redirectUrl := "/workflows"

	users, err := FindGeneratedUser(ctx, strings.ToLower(strings.TrimSpace(userName)))
	if err == nil && len(users) > 0 {
		for _, user := range users {
			log.Printf("%s - %s", user.GeneratedUsername, userName)
			if user.GeneratedUsername == userName {
				log.Printf("[AUDIT] Found user %s (%s) which matches SSO info for %s. Redirecting to login!", user.Username, user.Id, userName)

				//log.Printf("SESSION: %s", user.Session)

				expiration := time.Now().Add(3600 * time.Second)
				//if len(user.Session) == 0 {
				log.Printf("[INFO] User does NOT have session - creating")
				sessionToken := uuid.NewV4().String()

				newCookie := http.Cookie{
					Name:    "session_token",
					Value:   sessionToken,
					Expires: expiration,
				}

				if project.Environment == "cloud" {
					newCookie.Domain = ".shuffler.io"
					newCookie.Secure = true
					newCookie.HttpOnly = true
				}

				http.SetCookie(resp, &newCookie)

				newCookie.Name = "__session"
				http.SetCookie(resp, &newCookie)

				err = SetSession(ctx, user, sessionToken)
				if err != nil {
					log.Printf("[WARNING] Error creating session for user: %s", err)
					resp.WriteHeader(401)
					resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed setting session"}`)))
					return
				}

				user.Session = sessionToken
				err = SetUser(ctx, &user, false)
				if err != nil {
					log.Printf("[WARNING] Failed updating user when setting session: %s", err)
					resp.WriteHeader(401)
					resp.Write([]byte(`{"success": false, "reason": "Failed user update during session storage (2)"}`))
					return
				}

				//redirectUrl = fmt.Sprintf("%s?source=SSO&id=%s", redirectUrl, session)
				http.Redirect(resp, request, redirectUrl, http.StatusSeeOther)
				return
			}
		}
	}

	// Normal user. Checking because of backwards compatibility. Shouldn't break anything as we have unique names
	users, err = FindUser(ctx, strings.ToLower(strings.TrimSpace(userName)))
	if err == nil && len(users) > 0 {
		for _, user := range users {
			if user.Username == userName {
				log.Printf("[AUDIT] Found user %s (%s) which matches SSO info for %s. Redirecting to login %s!", user.Username, user.Id, userName, redirectUrl)

				//log.Printf("SESSION: %s", user.Session)

				expiration := time.Now().Add(3600 * time.Second)
				//if len(user.Session) == 0 {
				log.Printf("[INFO] User does NOT have session - creating")
				sessionToken := uuid.NewV4().String()
				newCookie := &http.Cookie{
					Name:    "session_token",
					Value:   sessionToken,
					Expires: expiration,
				}

				if project.Environment == "cloud" {
					newCookie.Domain = ".shuffler.io"
					newCookie.Secure = true
					newCookie.HttpOnly = true
				}

				http.SetCookie(resp, newCookie)

				newCookie.Name = "__session"
				http.SetCookie(resp, newCookie)

				err = SetSession(ctx, user, sessionToken)
				if err != nil {
					log.Printf("[WARNING] Error creating session for user: %s", err)
					resp.WriteHeader(401)
					resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed setting session"}`)))
					return
				}

				user.Session = sessionToken
				err = SetUser(ctx, &user, false)
				if err != nil {
					log.Printf("[WARNING] Failed updating user when setting session: %s", err)
					resp.WriteHeader(401)
					resp.Write([]byte(`{"success": false, "reason": "Failed user update during session storage (2)"}`))
					return
				}

				//redirectUrl = fmt.Sprintf("%s?source=SSO&id=%s", redirectUrl, session)
				http.Redirect(resp, request, redirectUrl, http.StatusSeeOther)
				return
			}
		}
	}

	/*
		orgs, err := GetAllOrgs(ctx)
		if err != nil {
			log.Printf("[WARNING] Failed finding orgs during SSO setup: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed getting valid organizations"}`)))
			return
		}

		foundOrg := Org{}
		for _, org := range orgs {
			if len(org.ManagerOrgs) == 0 {
				foundOrg = org
				break
			}
		}
	*/

	if len(org.Id) == 0 {
		log.Printf("[WARNING] Failed finding a valid org (default) without suborgs during SSO setup")
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed finding valid SSO auto org"}`)))
		return
	}

	log.Printf("[AUDIT] Adding user %s to org %s (%s) through single sign-on", userName, org.Name, org.Id)
	newUser := new(User)
	// Random password to ensure its not empty
	newUser.Password = uuid.NewV4().String()
	newUser.Username = userName
	newUser.GeneratedUsername = userName
	newUser.Verified = true
	newUser.Active = true
	newUser.CreationTime = time.Now().Unix()
	newUser.Orgs = []string{org.Id}
	newUser.LoginType = "OpenID"
	newUser.Role = "user"
	newUser.Session = uuid.NewV4().String()

	verifyToken := uuid.NewV4()
	ID := uuid.NewV4()
	newUser.Id = ID.String()
	newUser.VerificationToken = verifyToken.String()

	expiration := time.Now().Add(3600 * time.Second)
	//if len(user.Session) == 0 {
	log.Printf("[INFO] User does NOT have session - creating")
	sessionToken := uuid.NewV4().String()

	newCookie := &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiration,
	}

	if project.Environment == "cloud" {
		newCookie.Domain = ".shuffler.io"
		newCookie.Secure = true
		newCookie.HttpOnly = true
	}

	http.SetCookie(resp, newCookie)

	newCookie.Name = "__session"
	http.SetCookie(resp, newCookie)

	err = SetSession(ctx, *newUser, sessionToken)
	if err != nil {
		log.Printf("[WARNING] Error creating session for user: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed setting session"}`)))
		return
	}

	newUser.Session = sessionToken
	err = SetUser(ctx, newUser, true)
	if err != nil {
		log.Printf("[WARNING] Failed setting new user in DB: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed updating the user"}`)))
		return
	}

	http.Redirect(resp, request, redirectUrl, http.StatusSeeOther)
	return
}

// Example implementation of SSO, including a redirect for the user etc
// Should make this stuff only possible after login
func HandleSSO(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	//log.Printf("SSO LOGIN: %s", request)
	// Deserialize
	// Serialize

	// SAML
	//entryPoint := "https://dev-23367303.okta.com/app/dev-23367303_shuffletest_1/exk1vg1j7bYUYEG0k5d7/sso/saml"
	redirectUrl := "http://localhost:3001/workflows"
	backendUrl := os.Getenv("SSO_REDIRECT_URL")
	if len(backendUrl) == 0 && len(os.Getenv("BASE_URL")) > 0 {
		backendUrl = os.Getenv("BASE_URL")
	}

	if len(backendUrl) > 0 {
		redirectUrl = fmt.Sprintf("%s/workflows", backendUrl)
	}

	if project.Environment == "cloud" {
		redirectUrl = "https://shuffler.io/workflows"

		if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
			backendUrl = fmt.Sprintf("https://%s.%s.r.appspot.com/workflows", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
		}

		if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
			backendUrl = fmt.Sprintf("%s/workflows", os.Getenv("SHUFFLE_CLOUDRUN_URL"))
		}
	}

	log.Printf("[DEBUG] Using %s as redirectUrl in SSO", backendUrl)

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Error with body read of SSO: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Parsing out without using Field
	// This is a mess, but all made to handle base64 and equal signs
	parsedSAML := ""
	for _, item := range strings.Split(string(body), "&") {
		//log.Printf("Got body with info: %s", item)
		if strings.Contains(item, "SAMLRequest") || strings.Contains(item, "SAMLResponse") {
			equalsplit := strings.Split(item, "=")
			addedEquals := len(equalsplit)
			if len(equalsplit) >= 2 {
				//bareEquals := strings.Join(equalsplit[1:len(equalsplit)-1], "=")
				bareEquals := equalsplit[1]
				//log.Printf("Equal: %s", bareEquals)
				_ = addedEquals
				//if len(strings.Split(bareEquals, "=")) < addedEquals {
				//	bareEquals += "="
				//}

				decodedValue, err := url.QueryUnescape(bareEquals)
				if err != nil {
					log.Printf("[WARNING] Failed url query escape: %s", err)
					resp.WriteHeader(401)
					resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed decoding saml value"}`)))
					return
				}

				if strings.Contains(decodedValue, " ") {
					decodedValue = strings.Replace(decodedValue, " ", "+", -1)
				}

				parsedSAML = decodedValue
				break
			}
		}
	}

	if len(parsedSAML) == 0 {
		log.Printf("[WARNING] No SAML to be parsed from request.")
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "No data to parse. Is the request missing the SAMLResponse query?"}`)))
		return
	}

	bytesXML, err := base64.StdEncoding.DecodeString(parsedSAML)
	if err != nil {
		log.Printf("[WARNING] Failed base64 decode of SAML: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed base64 decoding in SAML"}`)))
		return
	}

	//log.Printf("Parsed: %s", bytesXML)

	// Sample request in keycloak lab env
	// PS: Should it ever come this way..?
	//<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns="urn:oasis:names:tc:SAML:2.0:assertion" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" AssertionConsumerServiceURL="http://192.168.55.2:8080/auth/realms/ShuffleSSOSaml/broker/shaffuru/endpoint" Destination="http://192.168.55.2:3001/api/v1/login_sso" ForceAuthn="false" ID="" IssueInstant="2022-01-31T20:24:37.238Z" ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Version="2.0"><saml:Issuer>http://192.168.55.2:8080/auth/realms/ShuffleSSOSaml</saml:Issuer><samlp:NameIDPolicy AllowCreate="false" Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"/></samlp:AuthnRequest>

	var samlResp SAMLResponse
	err = xml.Unmarshal(bytesXML, &samlResp)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%s", err), "AuthnRequest") {
			var newSamlResp SamlRequest
			err = xml.Unmarshal(bytesXML, &newSamlResp)
			if err != nil {
				log.Printf("[WARNING] Failed XML unmarshal (2): %s", err)
				resp.WriteHeader(401)
				resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed XML unpacking in SAML (2)"}`)))
				return
			}

			// Being here means we need to redirect
			log.Printf("[DEBUG] Handling authnrequest redirect back to %s? That's not how any of this works.", newSamlResp.AssertionConsumerServiceURL)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed XML unpacking in SAML (2)"}`)))
			return

			// User tries to access a protected resource on the SP. SP checks if the user has a local (and authenticated session). If not it generates a SAML <AuthRequest> which includes a random id. The SP then redirects the user to the IDP with this AuthnRequest.

			return
		} else {
			log.Printf("[WARNING] Failed XML unmarshal: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed XML unpacking in SAML (1)"}`)))
			return
		}
	}

	baseCertificate := samlResp.Signature.KeyInfo.X509Data.X509Certificate
	if len(baseCertificate) == 0 {
		//log.Printf("%s", samlResp.Signature.KeyInfo.X509Data)
		baseCertificate = samlResp.Assertion.Signature.KeyInfo.X509Data.X509Certificate
	}

	//log.Printf("\n\n%d - CERT: %s\n\n", len(baseCertificate), baseCertificate)
	parsedX509Key := fixCertificate(baseCertificate)

	ctx := GetContext(request)
	matchingOrgs, err := GetOrgByField(ctx, "sso_config.sso_certificate", parsedX509Key)
	if err != nil && len(matchingOrgs) == 0 {
		log.Printf("[DEBUG] BYTES FROM REQUEST (DEBUG): %s", string(bytesXML))

		log.Printf("[WARNING] Bad certificate (%d): Failed to find a org with certificate matching the SSO", len(parsedX509Key))
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed finding an org with the right certificate"}`)))
		return
	}

	// Validating the orgs
	if len(matchingOrgs) >= 1 {
		newOrgs := []Org{}
		for _, org := range matchingOrgs {
			if org.SSOConfig.SSOCertificate == parsedX509Key {
				newOrgs = append(newOrgs, org)
			} else {
				log.Printf("[WARNING] Skipping org append because bad cert: %d vs %d", len(org.SSOConfig.SSOCertificate), len(parsedX509Key))

				//log.Printf(parsedX509Key)
				//log.Printf(org.SSOConfig.SSOCertificate)
			}
		}

		matchingOrgs = newOrgs
	}

	if len(matchingOrgs) != 1 {
		log.Printf("[DEBUG] BYTES FROM REQUEST (2 - DEBUG): %s", string(bytesXML))
		log.Printf("[WARNING] Bad certificate (%d). Original orgs: %d: X509 doesnt match certificate for any organization", len(parsedX509Key), len(matchingOrgs))
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Certificate for SSO doesn't match any organization"}`)))
		return
	}

	foundOrg := matchingOrgs[0]
	userName := strings.ToLower(strings.TrimSpace(samlResp.Assertion.Subject.NameID.Text))
	if !strings.Contains(userName, "@") {
		log.Printf("[ERROR] Bad username, but allowing due to SSO: %s. Full Subject: %#v", userName, samlResp.Assertion.Subject)
	}

	if len(userName) == 0 {
		log.Printf("[WARNING] Failed finding user - No name: %s", samlResp.Assertion.Subject)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed finding a user to authenticate"}`)))
		return
	}

	// Start actually fixing the user
	// 1. Check if the user exists - if it does - give it a valid cookie
	// 2. If it doesn't, find the correct org to connect them with, then register them

	/*
		if project.Environment == "cloud" {
			log.Printf("[WARNING] SAML SSO is not implemented for cloud yet")
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Cloud SSO is not available for you"}`)))
			return
		}
	*/

	users, err := FindGeneratedUser(ctx, strings.ToLower(strings.TrimSpace(userName)))
	if err == nil && len(users) > 0 {
		for _, user := range users {
			log.Printf("%s - %s", user.GeneratedUsername, userName)
			if user.GeneratedUsername == userName {
				log.Printf("[AUDIT] Found user %s (%s) which matches SSO info for %s. Redirecting to login!", user.Username, user.Id, userName)

				if project.Environment == "cloud" {
					user.ActiveOrg.Id = matchingOrgs[0].Id

					DeleteCache(ctx, fmt.Sprintf("%s_workflows", user.Id))
					DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
					DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
					DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))
				}

				//log.Printf("SESSION: %s", user.Session)

				expiration := time.Now().Add(3600 * time.Second)
				//if len(user.Session) == 0 {
				log.Printf("[INFO] User does NOT have session - creating")
				sessionToken := uuid.NewV4().String()
				newCookie := &http.Cookie{
					Name:    "session_token",
					Value:   sessionToken,
					Expires: expiration,
				}

				if project.Environment == "cloud" {
					newCookie.Domain = ".shuffler.io"
					newCookie.Secure = true
					newCookie.HttpOnly = true
				}

				http.SetCookie(resp, newCookie)

				newCookie.Name = "__session"
				http.SetCookie(resp, newCookie)

				err = SetSession(ctx, user, sessionToken)
				if err != nil {
					log.Printf("[WARNING] Error creating session for user: %s", err)
					resp.WriteHeader(401)
					resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed setting session"}`)))
					return
				}

				user.Session = sessionToken
				err = SetUser(ctx, &user, false)
				if err != nil {
					log.Printf("[WARNING] Failed updating user when setting session: %s", err)
					resp.WriteHeader(401)
					resp.Write([]byte(`{"success": false, "reason": "Failed user update during session storage (2)"}`))
					return
				}

				//redirectUrl = fmt.Sprintf("%s?source=SSO&id=%s", redirectUrl, session)
				http.Redirect(resp, request, redirectUrl, http.StatusSeeOther)
				return
			}
		}
	}

	// Normal user. Checking because of backwards compatibility. Shouldn't break anything as we have unique names
	users, err = FindUser(ctx, strings.ToLower(strings.TrimSpace(userName)))
	if err == nil && len(users) > 0 {
		for _, user := range users {
			if user.Username == userName {
				log.Printf("[AUDIT] Found user %s (%s) which matches SSO info for %s. Redirecting to login %s!", user.Username, user.Id, userName, redirectUrl)

				//log.Printf("SESSION: %s", user.Session)
				if project.Environment == "cloud" {
					user.ActiveOrg.Id = matchingOrgs[0].Id
				}

				expiration := time.Now().Add(3600 * time.Second)
				//if len(user.Session) == 0 {
				log.Printf("[INFO] User does NOT have session - creating")
				sessionToken := uuid.NewV4().String()
				newCookie := &http.Cookie{
					Name:    "session_token",
					Value:   sessionToken,
					Expires: expiration,
				}

				if project.Environment == "cloud" {
					newCookie.Domain = ".shuffler.io"
					newCookie.Secure = true
					newCookie.HttpOnly = true
				}

				http.SetCookie(resp, newCookie)

				newCookie.Name = "__session"
				http.SetCookie(resp, newCookie)

				err = SetSession(ctx, user, sessionToken)
				if err != nil {
					log.Printf("[WARNING] Error creating session for user: %s", err)
					resp.WriteHeader(401)
					resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed setting session"}`)))
					return
				}

				user.Session = sessionToken
				err = SetUser(ctx, &user, false)
				if err != nil {
					log.Printf("[WARNING] Failed updating user when setting session: %s", err)
					resp.WriteHeader(401)
					resp.Write([]byte(`{"success": false, "reason": "Failed user update during session storage (2)"}`))
					return
				}

				//redirectUrl = fmt.Sprintf("%s?source=SSO&id=%s", redirectUrl, session)
				http.Redirect(resp, request, redirectUrl, http.StatusSeeOther)
				return
			}
		}
	}

	/*
		orgs, err := GetAllOrgs(ctx)
		if err != nil {
			log.Printf("[WARNING] Failed finding orgs during SSO setup: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed getting valid organizations"}`)))
			return
		}

		foundOrg := Org{}
		for _, org := range orgs {
			if len(org.ManagerOrgs) == 0 {
				foundOrg = org
				break
			}
		}
	*/

	if len(foundOrg.Id) == 0 {
		log.Printf("[WARNING] Failed finding a valid org (default) without suborgs during SSO setup")
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed finding valid SSO auto org"}`)))
		return
	}

	log.Printf("[AUDIT] Adding user %s to org %s (%s) through single sign-on", userName, foundOrg.Name, foundOrg.Id)
	newUser := new(User)
	// Random password to ensure its not empty
	newUser.Password = uuid.NewV4().String()
	newUser.Username = userName
	newUser.GeneratedUsername = userName
	newUser.Verified = true
	newUser.Active = true
	newUser.CreationTime = time.Now().Unix()
	newUser.Orgs = []string{foundOrg.Id}
	newUser.LoginType = "SSO"
	newUser.Role = "user"
	newUser.Session = uuid.NewV4().String()

	newUser.ActiveOrg.Id = matchingOrgs[0].Id

	verifyToken := uuid.NewV4()
	ID := uuid.NewV4()
	newUser.Id = ID.String()
	newUser.VerificationToken = verifyToken.String()

	expiration := time.Now().Add(3600 * time.Second)
	//if len(user.Session) == 0 {
	log.Printf("[INFO] User does NOT have session - creating")
	sessionToken := uuid.NewV4().String()
	newCookie := &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiration,
	}

	if project.Environment == "cloud" {
		newCookie.Domain = ".shuffler.io"
		newCookie.Secure = true
		newCookie.HttpOnly = true
	}

	http.SetCookie(resp, newCookie)

	newCookie.Name = "__session"
	http.SetCookie(resp, newCookie)

	err = SetSession(ctx, *newUser, sessionToken)
	if err != nil {
		log.Printf("[WARNING] Error creating session for user: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed setting session"}`)))
		return
	}

	newUser.Session = sessionToken
	err = SetUser(ctx, newUser, true)
	if err != nil {
		log.Printf("[WARNING] Failed setting new user in DB: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed updating the user"}`)))
		return
	}

	http.Redirect(resp, request, redirectUrl, http.StatusSeeOther)
	return
}

// Downloads documentation from Github to be placed in an app/workflow as markdown
// Caching no matter what, with no retries
func DownloadFromUrl(ctx context.Context, url string) ([]byte, error) {
	cacheKey := fmt.Sprintf("docs_%s", url)
	cache, err := GetCache(ctx, cacheKey)
	if err == nil {
		cacheData := []byte(cache.([]uint8))
		return cacheData, nil
	}

	httpClient := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)

	if err != nil {
		SetCache(ctx, cacheKey, []byte{}, 30)
		return []byte{}, err
	}

	newresp, err := httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}

	//log.Printf("URL %s, RESP: %d", url, newresp.StatusCode)
	if newresp.StatusCode != 200 {
		SetCache(ctx, cacheKey, []byte{}, 30)

		return []byte{}, errors.New(fmt.Sprintf("No body to handle for %s. Status: %d", url, newresp.StatusCode))
	}

	body, err := ioutil.ReadAll(newresp.Body)
	if err != nil {
		SetCache(ctx, cacheKey, []byte{}, 30)
		return []byte{}, err
	}

	//log.Printf("Documentation: %s", string(body))
	if len(body) > 0 {
		err = SetCache(ctx, cacheKey, body, 30)
		if err != nil {
			log.Printf("[WARNING] Failed setting cache for workflow/app doc %s: %s", url, err)
		}
		return body, nil
	}

	SetCache(ctx, cacheKey, []byte{}, 30)
	return []byte{}, errors.New(fmt.Sprintf("No body to handle for %s", url))
}

// // New execution with firestore
func PrepareWorkflowExecution(ctx context.Context, workflow Workflow, request *http.Request, maxExecutionDepth int64) (WorkflowExecution, ExecInfo, string, error) {
	log.Printf("[INFO] At start of prepare exec")

	workflowBytes, err := json.Marshal(workflow)
	if err != nil {
		log.Printf("[WARNING] Failed workflow unmarshal in execution: %s", err)
		return WorkflowExecution{}, ExecInfo{}, "", err
	}

	//log.Printf(workflow)
	var workflowExecution WorkflowExecution
	err = json.Unmarshal(workflowBytes, &workflowExecution.Workflow)
	if err != nil {
		log.Printf("[WARNING] Failed prepare execution unmarshaling: %s", err)
		return WorkflowExecution{}, ExecInfo{}, "Failed unmarshal during execution", err
	}

	if len(workflow.OrgId) > 0 {
		workflowExecution.ExecutionOrg = workflow.OrgId
		workflowExecution.OrgId = workflow.OrgId
	}

	log.Printf("[INFO] Checking request methods and such")

	makeNew := true
	start, startok := request.URL.Query()["start"]
	if request.Method == "POST" {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Printf("[ERROR] Failed request POST read: %s", err)
			return WorkflowExecution{}, ExecInfo{}, "Failed getting body", err
		}

		// This one doesn't really matter.
		log.Printf("[INFO] Running POST execution with body of length %d for workflow %s", len(string(body)), workflowExecution.Workflow.ID)

		if len(body) >= 4 {
			if body[0] == 34 && body[len(body)-1] == 34 {
				body = body[1 : len(body)-1]
			}
			if body[0] == 34 && body[len(body)-1] == 34 {
				body = body[1 : len(body)-1]
			}
		}

		sourceAuth, sourceAuthOk := request.URL.Query()["source_auth"]
		if sourceAuthOk {
			//log.Printf("\n\n\nSETTING SOURCE WORKFLOW AUTH TO %s!!!\n\n\n", sourceAuth[0])
			workflowExecution.ExecutionSourceAuth = sourceAuth[0]
		} else {
			//log.Printf("Did NOT get source workflow")
		}

		sourceNode, sourceNodeOk := request.URL.Query()["source_node"]
		if sourceNodeOk {
			//log.Printf("\n\n\nSETTING SOURCE WORKFLOW NODE TO %s!!!\n\n\n", sourceNode[0])
			workflowExecution.ExecutionSourceNode = sourceNode[0]
		} else {
			//log.Printf("Did NOT get source workflow")
		}

		//workflowExecution.ExecutionSource = "default"
		sourceWorkflow, sourceWorkflowOk := request.URL.Query()["source_workflow"]
		if sourceWorkflowOk {
			//log.Printf("Got source workflow %s", sourceWorkflow)
			workflowExecution.ExecutionSource = sourceWorkflow[0]
		} else {
			//log.Printf("Did NOT get source workflow")
		}

		sourceExecution, sourceExecutionOk := request.URL.Query()["source_execution"]
		parentExecution := &WorkflowExecution{}
		if sourceExecutionOk {
			//log.Printf("[INFO] Got source execution%s", sourceExecution)
			workflowExecution.ExecutionParent = sourceExecution[0]

			// FIXME: Get the execution and check count
			//workflowExecution.SubExecutionCount += 1

			//log.Printf("\n\n[INFO] PARENT!!: %s\n\n", workflowExecution.ExecutionParent)
			parentExecution, err = GetWorkflowExecution(ctx, workflowExecution.ExecutionParent)
			if err == nil {
				workflowExecution.SubExecutionCount = parentExecution.SubExecutionCount + 1
			}

			// Subflow are JUST lower than manual executions
			if workflowExecution.Priority == 0 {
				workflowExecution.Priority = 9
			}
		} else {
			//log.Printf("Did NOT get source execution")
		}

		// Checks whether the subflow has been ran before based on parent execution ID + parent execution node ID (always unique)
		// Used to deduplicate runs
		if len(workflowExecution.ExecutionParent) > 0 && len(workflowExecution.ExecutionSourceNode) > 0 {
			// Check if it should be looping:
			// 1. Get workflowExecution.ExecutionParent's workflow
			// 2. Find the ExecutionSourceNode
			// 3. Check if the value of it is looping
			var parentErr error
			if len(parentExecution.ExecutionId) == 0 {
				parentExecution, parentErr = GetWorkflowExecution(ctx, workflowExecution.ExecutionParent)
			}

			allowContinuation := false
			if parentErr == nil {
				found := false
				for _, trigger := range parentExecution.Workflow.Triggers {
					if trigger.ID != workflowExecution.ExecutionSourceNode {
						continue
					}

					found = true

					//$Get_Offenses.# -> Allow to run more
					for _, param := range trigger.Parameters {
						if param.Name == "argument" {
							if strings.Contains(param.Value, "$") && strings.Contains(param.Value, ".#") {
								allowContinuation = true
								break
							}
						}
					}

					if allowContinuation {
						break
					}
				}

				if !found {
					// Added from subflow trigger -> action translation
					for _, action := range parentExecution.Workflow.Actions {
						if action.ID != workflowExecution.ExecutionSourceNode {
							continue
						}

						found = true

						//$Get_Offenses.# -> Allow to run more
						for _, param := range action.Parameters {
							if param.Name == "argument" {
								if strings.Contains(param.Value, "$") && strings.Contains(param.Value, ".#") {
									allowContinuation = true
									break
								}
							}
						}

						if allowContinuation {
							break
						}
					}
				}
			}

			if allowContinuation == false {
				newExecId := fmt.Sprintf("%s_%s_%s", workflowExecution.ExecutionParent, workflowExecution.ExecutionId, workflowExecution.ExecutionSourceNode)
				cache, err := GetCache(ctx, newExecId)
				if err == nil {
					cacheData := []byte(cache.([]uint8))

					newexec := WorkflowExecution{}
					log.Printf("[WARNING] Subflow exec %s already found - returning", newExecId)

					// Returning to be used in worker
					err = json.Unmarshal(cacheData, &newexec)
					if err == nil {
						return newexec, ExecInfo{}, fmt.Sprintf("Subflow for %s has already been executed", newExecId), errors.New(fmt.Sprintf("Subflow for %s has already been executed", newExecId))
					}

					return workflowExecution, ExecInfo{}, fmt.Sprintf("Subflow for %s has already been executed", newExecId), errors.New(fmt.Sprintf("Subflow for %s has already been executed", newExecId))
				}

				cacheData := []byte("1")
				err = SetCache(ctx, newExecId, cacheData, 2)
				if err != nil {
					log.Printf("[WARNING] Failed setting cache for action %s: %s", newExecId, err)
				} else {
				}
			}
		}

		if len(string(body)) < 50 {
			log.Printf("[DEBUG][%s] Body: %s", workflowExecution.ExecutionId, string(body))
		} else {
			// Here for debug purposes
			log.Printf("[DEBUG][%s] Body len: %d", workflowExecution.ExecutionId, len(string(body)))
		}

		var execution ExecutionRequest
		err = json.Unmarshal(body, &execution)
		if err != nil {
			//log.Printf("[WARNING] Failed execution POST unmarshaling - continuing anyway: %s", err)
			//return WorkflowExecution{}, "", err
		}

		// Ensuring it works even if startpoint isn't defined
		if execution.Start == "" && len(body) > 0 && len(execution.ExecutionSource) == 0 {
			execution.ExecutionArgument = string(body)
		}

		// FIXME - this should have "execution_argument" from executeWorkflow frontend
		//log.Printf("EXEC: %s", execution)
		if len(execution.ExecutionArgument) > 0 {
			workflowExecution.ExecutionArgument = execution.ExecutionArgument
		}

		if len(execution.ExecutionSource) > 0 {
			workflowExecution.ExecutionSource = execution.ExecutionSource

			if workflowExecution.Priority == 0 {
				workflowExecution.Priority = 5
			}
		}

		//log.Printf("Execution data: %s", execution)
		if len(execution.Start) == 36 && len(workflow.Actions) > 0 {
			log.Printf("[INFO] Should start execution on node %s", execution.Start)
			workflowExecution.Start = execution.Start

			found := false
			newStartnode := ""
			for _, action := range workflow.Actions {
				if action.ID == execution.Start {
					found = true
					break
				}

				if action.IsStartNode {
					newStartnode = action.ID
				}
			}

			if !found {
				if len(newStartnode) > 0 {
					execution.Start = newStartnode
				} else {
					log.Printf("[ERROR] Action %s was NOT found, and no other startnode found! Exiting execution.", execution.Start)
					return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Startnode %s was not found in actions", workflow.Start), errors.New(fmt.Sprintf("Startnode %s was not found in actions", workflow.Start))
				}
			}
		} else if len(execution.Start) > 0 {
			//return WorkflowExecution{}, fmt.Sprintf("Startnode %s was not found in actions", execution.Start), errors.New(fmt.Sprintf("Startnode %s was not found in actions", execution.Start))
		}

		if len(execution.ExecutionId) == 36 {
			workflowExecution.ExecutionId = execution.ExecutionId
		} else {
			sessionToken := uuid.NewV4()
			workflowExecution.ExecutionId = sessionToken.String()
		}
	} else {
		// Check for parameters of start and ExecutionId
		// This is mostly used for user input trigger
		answer, answerok := request.URL.Query()["answer"]
		referenceId, referenceok := request.URL.Query()["reference_execution"]
		authorization, authorizationok := request.URL.Query()["authorization"]
		if answerok && referenceok && authorizationok {
			// If answer is false, reference execution with result
			log.Printf("[INFO] Should update reference execution and return, no need for further execution! exec ref: %s. Auth: %s", referenceId[0], authorization[0])

			// Get the reference execution
			oldExecution, err := GetWorkflowExecution(ctx, referenceId[0])
			if err != nil {
				log.Printf("[INFO] Failed getting execution (execution) %s: %s", referenceId[0], err)
				return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Failed getting execution ID %s because it doesn't exist.", referenceId[0]), err
			}

			log.Printf("[INFO] Got execution %s. Workflow: %s", oldExecution.ExecutionId, oldExecution.Workflow.Name)

			if oldExecution.Workflow.ID != workflow.ID {
				log.Printf("[INFO] Wrong workflowid!")
				return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Bad workflow ID in get %s", referenceId), errors.New("Bad workflow ID")
			}

			if authorization[0] != oldExecution.Authorization {
				log.Printf("[AUDIT][%s] Wrong authorization for execution during userinput! %s vs %s", referenceId[0], authorization[0], oldExecution.Authorization)
				return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Bad authorization in get %s", referenceId), errors.New("Bad authorization key")
			}

			if len(start) == 0 {
				// Just guessing~
				for _, trigger := range workflow.Triggers {
					if trigger.AppName == "User Input" {
						start = []string{trigger.ID}
						break
					}
				}
			}

			//log.Printf("Result len: %d", len(oldExecution.Results))
			newResults := []ActionResult{}
			foundresult := ActionResult{}
			for _, result := range oldExecution.Results {
				log.Printf("Action: %s - %s", result.Action.ID, start[0])
				if result.Status != "WAITING" {
					log.Printf("[WARNING] '%s' has status '%s'", result.Action.Label, result.Status)
					// Testing
					if result.Action.Label == "User_Input_1" {
						result.Status = "WAITING"
					}
				}

				if result.Status == "WAITING" {
					log.Printf("[INFO] Found result: %s (%s)", result.Action.Label, result.Action.ID)

					var userinputResp UserInputResponse
					err = json.Unmarshal([]byte(result.Result), &userinputResp)
					// Error here should just be warnings
					if err != nil {
						log.Printf("\n\n[DEBUG] Failed unmarshalling userinput (not critical): %s\n\n", err)
					}

					//if err == nil {
					userinputResp.ClickInfo.Clicked = true
					userinputResp.ClickInfo.Time = time.Now().Unix()
					userinputResp.ClickInfo.IP = request.RemoteAddr
					userinputResp.ClickInfo.Note = ""

					user, err := HandleApiAuthentication(nil, request)
					if err == nil && user.Username != "" {
						userinputResp.ClickInfo.User = user.Username
					}

					b, err := json.Marshal(userinputResp)
					if err != nil {
						log.Printf("[ERROR] Failed marshalling userinput: %s", err)
					} else {
						result.Result = string(b)
					}

					result.CompletedAt = int64(time.Now().Unix())
					if answer[0] == "false" {
						result.Status = "ABORTED"
					} else {
						result.Status = "SUCCESS"
					}

					// Should send result to self?
					/*
						fullMarshal, err := json.Marshal(result)
						if err == nil {
							// Should set  cache for it
							actionCacheId := fmt.Sprintf("%s_%s_result", result.ExecutionId, result.Action.ID)
							err = SetCache(ctx, actionCacheId, fullMarshal, 35)
							if err != nil {
								log.Printf("[ERROR] Failed setting cache for action result %s: %s", actionCacheId, err)
							}
						} else {
							log.Printf("[ERROR] Failed marshalling user input WAITING result: %s", err)
						}
					*/

					/*
						} else {
							log.Printf("[ERROR] Failed unmarshalling userinput: %s", err)

							// FIXME: Add note functionality again
							//note, noteok := request.URL.Query()["note"]
							//if noteok {
							//	result.Result = fmt.Sprintf("User note: %s", note[0])
							//} else {
							//	result.Result = fmt.Sprintf("User clicked %s", answer[0])
							//}

							// Stopping the whole thing
							if answer[0] == "false" {
								result.CompletedAt = int64(time.Now().Unix())
								result.Status = "ABORTED"

								// Abort from here?
								oldExecution.Status = result.Status
								oldExecution.Result = result.Result
								oldExecution.LastNode = result.Action.ID
							} else {
								result.CompletedAt = int64(time.Now().Unix())
								result.Status = "SUCCESS"
								result.Result = `{"success": true, "reason": "Continuing from user input", "user_input": true}`
							}
						}
					*/

					foundresult = result
					newResults = append(newResults, result)
				} else {
					newResults = append(newResults, result)
				}
			}

			if foundresult.Action.AppName != "" {
				// Should resend the result to redeploy the job?
				log.Printf("[INFO][%s] Rerunning node for user input with WAITING", oldExecution.ExecutionId)
				b, err := json.Marshal(foundresult)
				if err != nil {
					log.Printf("[WARNING][%s] Failed to run node for user input with WAITING", oldExecution.ExecutionId)
				} else {
					ResendActionResult(b, 4)
				}
			} else {
				log.Printf("[WARNING][%s] No job to rerun for user input as a WAITING node was not found", oldExecution.ExecutionId)
			}

			//if err != nil {
			/*
				oldExecution.Results = newResults
				err = SetWorkflowExecution(ctx, *oldExecution, true)
				if err != nil {
					log.Printf("[WARNING] Error saving workflow execution actionresult setting: %s", err)
					return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Failed setting workflowexecution actionresult in execution: %s", err), err
				}

				if foundresult.Action.AppName != "" {
					b, err := json.Marshal(foundresult)
					if err != nil {
						log.Printf("[ERROR] Error marshalling actionresult: %s", err)
					} else {
						log.Printf("[INFO] Sending user input result: %s", string(b))
						ResendActionResult(b, 4)
					}
				} else {
					log.Printf("[INFO] No WAITING node found")
					return *oldExecution, ExecInfo{}, "", errors.New("User Input: Already finished")
				}
			*/

			// Add new execution to queue?
			//if os.Getenv("SHUFFLE_SWARM_CONFIG") == "run" && (project.Environment == "" || project.Environment == "worker") {

			return *oldExecution, ExecInfo{}, "", errors.New("User Input")
		}

		if referenceok {
			log.Printf("[DEBUG] Handling an old execution continuation! Start: %s", start)

			// Will use the old name, but still continue with NEW ID
			oldExecution, err := GetWorkflowExecution(ctx, referenceId[0])
			if err != nil {
				log.Printf("[ERROR] Failed getting execution (execution) %s: %s", referenceId[0], err)
				return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Failed getting execution ID %s because it doesn't exist.", referenceId[0]), err
			}

			if oldExecution.Status != "WAITING" {
				return WorkflowExecution{}, ExecInfo{}, "", errors.New("Workflow is no longer with status waiting. Can't continue.")
			}

			if startok {
				for _, result := range oldExecution.Results {
					if result.Action.ID == start[0] {
						if result.Status == "SUCCESS" || result.Status == "FINISHED" {
							// Disabling this to allow multiple continuations
							//return WorkflowExecution{}, ExecInfo{}, "", errors.New("This workflow has already been continued")
						}
						//log.Printf("Start: %s", result.Status)
					}
				}
			}

			workflowExecution = *oldExecution

			// A previously stopped workflow. Same priority as subflow.
			workflowExecution.Priority = 9
		}

		if len(workflowExecution.ExecutionId) == 0 {
			sessionToken := uuid.NewV4()
			workflowExecution.ExecutionId = sessionToken.String()
		} else {
			log.Printf("[DEBUG] Using the same executionId as before: %s", workflowExecution.ExecutionId)
			makeNew = false
		}

		// Don't override workflow defaults
	}

	if workflowExecution.SubExecutionCount == 0 {
		workflowExecution.SubExecutionCount = 1
	}

	//log.Printf("\n\nExecution count: %d", workflowExecution.SubExecutionCount)
	if workflowExecution.SubExecutionCount >= maxExecutionDepth {
		return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Max subflow of %d reached"), err
	}

	if workflowExecution.Priority == 0 {
		//log.Printf("\n\n[DEBUG] Set priority to 10 as it's manual?\n\n")
		workflowExecution.Priority = 10
	}

	if startok {
		//log.Printf("\n\n[INFO] Setting start to %s based on query!\n\n", start[0])
		//workflowExecution.Workflow.Start = start[0]
		workflowExecution.Start = start[0]
	}

	// FIXME - regex uuid, and check if already exists?
	if len(workflowExecution.ExecutionId) != 36 {
		log.Printf("Invalid uuid: %s", workflowExecution.ExecutionId)
		return WorkflowExecution{}, ExecInfo{}, "Invalid uuid", err
	}

	// FIXME - find owner of workflow
	// FIXME - get the actual workflow itself and build the request
	// MAYBE: Don't send the workflow within the pubsub, as this requires more data to be sent
	// Check if a worker already exists for company, else run one with:
	// locations, project IDs and subscription names

	// When app is executed:
	// Should update with status execution (somewhere), which will trigger the next node
	// IF action.type == internal, we need the internal watcher to be running and executing
	// This essentially means the WORKER has to be the responsible party for new actions in the INTERNAL landscape
	// Results are ALWAYS posted back to cloud@execution_id?
	if makeNew {
		workflowExecution.Type = "workflow"
		//workflowExecution.Stream = "tmp"
		//workflowExecution.WorkflowQueue = "tmp"
		//workflowExecution.SubscriptionNameNodestream = "testcompany-nodestream"
		//workflowExecution.Locations = []string{"europe-west2"}
		//workflowExecution.ProjectId = gceProject
		workflowExecution.WorkflowId = workflow.ID
		workflowExecution.StartedAt = int64(time.Now().Unix())
		workflowExecution.CompletedAt = 0
		workflowExecution.Authorization = uuid.NewV4().String()

		// Status for the entire workflow.
		workflowExecution.Status = "EXECUTING"
	}

	if len(workflowExecution.ExecutionSource) == 0 {
		//log.Printf("[INFO] No execution source (trigger) specified. Setting to default")
		workflowExecution.ExecutionSource = "default"
	}

	log.Printf("[INFO] Execution source is '%s' for execution ID %s in workflow %s. Organization: %s", workflowExecution.ExecutionSource, workflowExecution.ExecutionId, workflowExecution.Workflow.ID, workflowExecution.OrgId)

	workflowExecution.ExecutionVariables = workflow.ExecutionVariables
	if len(workflowExecution.Start) == 0 && len(workflowExecution.Workflow.Start) > 0 {
		workflowExecution.Start = workflowExecution.Workflow.Start
	}

	startnodeFound := false
	newStartnode := ""
	for actionIndex, item := range workflowExecution.Workflow.Actions {
		if item.ID == workflowExecution.Start {
			startnodeFound = true
		}

		if item.IsStartNode {
			newStartnode = item.ID
		}

		// Fix names of parameters
		for paramIndex, param := range item.Parameters {

			// Added after problem with api-secret -> apisecret
			if strings.Contains(param.Description, "header") {
				continue
			}

			if strings.Contains(param.Description, "query") {
				continue
			}

			newName := GetValidParameters([]string{param.Name})
			if len(newName) > 0 {
				workflowExecution.Workflow.Actions[actionIndex].Parameters[paramIndex].Name = newName[0]
			}
		}
	}

	if !startnodeFound {
		log.Printf("[WARNING] Couldn't find startnode %s among %d actions. Remapping to %s", workflowExecution.Start, len(workflowExecution.Workflow.Actions), newStartnode)

		if len(newStartnode) > 0 {
			workflowExecution.Start = newStartnode
		} else {
			return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Startnode couldn't be found"), errors.New("Startnode isn't defined in this workflow..")
		}
	}

	childNodes := FindChildNodes(workflowExecution, workflowExecution.Start, []string{}, []string{})

	//topic := "workflows"
	startFound := false
	// FIXME - remove this?
	newActions := []Action{}
	defaultResults := []ActionResult{}

	if project.Environment == "cloud" {
		//apps, err := GetPrioritizedApps(ctx, user)
		//if err != nil {
		//	log.Printf("[WARNING] Error: Failed getting apps during setup: %s", err)
		//}
	}

	allAuths := []AppAuthenticationStorage{}
	for _, action := range workflowExecution.Workflow.Actions {
		//action.LargeImage = ""
		if action.ID == workflowExecution.Start {
			startFound = true
		}

		// Fill in apikey?
		if project.Environment == "cloud" {

			if (action.AppName == "Shuffle Tools" || action.AppName == "email") && action.Name == "send_email_shuffle" || action.Name == "send_sms_shuffle" {
				for paramKey, param := range action.Parameters {
					// Autoreplace in general, even if there is a key. Overwrite previous configs to ensure this becomes the norm. Frontend also matches.
					if param.Name == "apikey" {
						//log.Printf("Autoreplacing apikey")

						// This will be in cache after running once or twice AKA fast
						org, err := GetOrg(ctx, workflowExecution.Workflow.OrgId)
						if err != nil {
							log.Printf("[ERROR] Error getting org in APIkey replacement: %s", err)
							continue
						}

						// Make sure to find one that's belonging to the org
						// Picking random last user if

						backupApikey := ""
						for _, user := range org.Users {
							if len(user.ApiKey) == 0 {
								continue
							}

							if user.Role != "org-reader" {
								backupApikey = user.ApiKey
							}

							if len(user.Orgs) == 1 || user.ActiveOrg.Id == workflowExecution.Workflow.OrgId {
								//log.Printf("Choice: %s, %s - %s", user.Username, user.Id, user.ApiKey)
								action.Parameters[paramKey].Value = user.ApiKey
								break
							}
						}

						if len(action.Parameters[paramKey].Value) == 0 {
							log.Printf("[WARNING] No apikey user found. Picking first random user")
							action.Parameters[paramKey].Value = backupApikey
						}

						break
					}
				}
			}
		}

		if action.Environment == "" {
			return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Environment is not defined for %s", action.Name), errors.New("Environment not defined!")
		}

		if len(action.AuthenticationId) > 0 {
			//log.Printf("Action for

			if len(allAuths) == 0 {
				allAuths, err = GetAllWorkflowAppAuth(ctx, workflow.ExecutingOrg.Id)
				if err != nil {
					log.Printf("[ERROR] Api authentication failed in get all app auth for ID %s: %s", workflow.ExecutingOrg.Id, err)
					return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Api authentication failed in get all app auth for %s: %s", workflow.ExecutingOrg.Id, err), err
				}
			}

			curAuth := AppAuthenticationStorage{Id: ""}
			authIndex := -1
			for innerIndex, auth := range allAuths {
				if auth.Id == action.AuthenticationId {
					authIndex = innerIndex
					curAuth = auth
					break
				}
			}

			if len(curAuth.Id) == 0 {
				return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("App Auth ID %s doesn't exist for app '%s'. Please re-authenticate the app.", action.AuthenticationId, action.AppName), errors.New(fmt.Sprintf("App Auth ID %s doesn't exist for app '%s'. Please re-authenticate the app.", action.AuthenticationId, action.AppName))
			}

			if curAuth.Encrypted {
				setField := true
				newFields := []AuthenticationStore{}
				fieldLength := 0
				for _, field := range curAuth.Fields {
					parsedKey := fmt.Sprintf("%s_%d_%s_%s", curAuth.OrgId, curAuth.Created, curAuth.Label, field.Key)
					newValue, err := HandleKeyDecryption([]byte(field.Value), parsedKey)
					if err != nil {
						log.Printf("[ERROR] Failed decryption in org %s for %s: %s", curAuth.OrgId, field.Key, err)
						setField = false
						//fieldLength = 0
						break
					}

					// Remove / at end of urls
					// TYPICALLY shouldn't use them.
					if field.Key == "url" {
						//log.Printf("Value2 (%s): %s", field.Key, string(newValue))
						if strings.HasSuffix(string(newValue), "/") {
							newValue = []byte(string(newValue)[0 : len(newValue)-1])
						}

						//log.Printf("Value2 (%s): %s", field.Key, string(newValue))
					}

					fieldLength += len(newValue)
					field.Value = string(newValue)
					newFields = append(newFields, field)
				}

				// There is some Very weird bug that has caused encryption to sometimes be skipped.
				// This is a way to discover when this happens properly.
				// The problem happens about every 10.000~ decryption, which is still way too much.
				// By adding the full total, there should be no problem with this, seeing as lengths are added together
				fieldNames := ""
				for _, field := range curAuth.Fields {
					fieldNames += field.Key + ", "
				}

				if setField {
					curAuth.Fields = newFields

					//log.Printf("[DEBUG] Outer decryption debugging for %s. Auth: %s, Fields: %s. Length: %d", curAuth.OrgId, curAuth.Label, fieldNames, fieldLength)
				} else {
					log.Printf("[ERROR] Outer decryption debugging for %s. Auth: %s. Fields: %s. Length: %d", curAuth.OrgId, curAuth.Label, fieldNames, fieldLength)

				}
			} else {
				//log.Printf("[INFO] AUTH IS NOT ENCRYPTED - attempting auto-encrypting if key is set!")
				err = SetWorkflowAppAuthDatastore(ctx, curAuth, curAuth.Id)
				if err != nil {
					log.Printf("[WARNING] Failed running encryption during execution: %s", err)
				}
			}

			newParams := []WorkflowAppActionParameter{}
			if strings.ToLower(curAuth.Type) == "oauth2" {
				//log.Printf("[DEBUG] Should replace auth parameters (Oauth2)")

				runRefresh := false
				refreshUrl := ""
				for _, param := range curAuth.Fields {
					if param.Key == "expiration" {
						val, err := strconv.Atoi(param.Value)
						timeNow := int64(time.Now().Unix())
						if err == nil {
							//log.Printf("Checking expiration vs timenow: %d %d. Err: %s", timeNow, int64(val)+120, err)
							if timeNow >= int64(val)+120 {
								log.Printf("[DEBUG] Should run refresh of Oauth2 for authentication ID '%s'!!", curAuth.Id)
								runRefresh = true
							}

						}

						continue
					}

					if param.Key == "refresh_url" {
						refreshUrl = param.Value
						continue
					}

					if param.Key != "url" && param.Key != "access_token" {
						//log.Printf("Skipping key %s", param.Key)
						continue
					}

					newParams = append(newParams, WorkflowAppActionParameter{
						Name:  param.Key,
						Value: param.Value,
					})
				}

				runRefresh = true
				if runRefresh {
					user := User{
						Username: "refresh",
						ActiveOrg: OrgMini{
							Id: curAuth.OrgId,
						},
					}

					if len(refreshUrl) == 0 {
						log.Printf("[ERROR] No Oauth2 request to run, as no refresh url is set!")
					} else {
						//log.Printf("[INFO] Running Oauth2 request with URL %s", refreshUrl)

						newAuth, err := RunOauth2Request(ctx, user, curAuth, true)
						if err != nil {
							log.Printf("[ERROR] Failed running oauth request to refresh oauth2 tokens: %s. Stopping Oauth2 continuation and sending abort for app. This is NOT critical, but means refreshing access_token failed, and it will stop working in the future.", err)
							// Commented out as we don't want to stop the app, but just continue with the old tokens
							/*
								actionRes := ActionResult{
									Action:        action,
									ExecutionId:   workflowExecution.ExecutionId,
									Authorization: workflowExecution.Authorization,
									Result:        fmt.Sprintf(`{"success": false, "reason": "Failed running oauth2 request to refresh oauth2 tokens for this app."}`),
									StartedAt:     workflowExecution.StartedAt,
									CompletedAt:   workflowExecution.StartedAt,
									Status:        "FAILURE",
								}

								workflowExecution.Results = append(workflowExecution.Results, actionRes)
								cacheData := []byte("1")

								newExecId := fmt.Sprintf("%s_%s", workflowExecution.ExecutionId, action.ID)
								err = SetCache(ctx, newExecId, cacheData, 2)
								if err != nil {
									log.Printf("[WARNING] Failed setting base cache for failed Oauth2 action %s: %s", newExecId, err)
								}

								b, err := json.Marshal(actionRes)
								if err == nil {
									err = SetCache(ctx, fmt.Sprintf("%s_result", newExecId), b, 2)
									if err != nil {
										log.Printf("[WARNING] Failed setting result cache for failed Oauth2 action %s: %s", newExecId, err)
									}
								}
							*/
						}

						log.Printf("[DEBUG] Setting new auth to index: %d and curauth", authIndex)
						allAuths[authIndex] = newAuth

						// Does the oauth2 replacement
						newParams = []WorkflowAppActionParameter{}
						for _, param := range newAuth.Fields {
							//log.Printf("FIELD: %s", param.Key, param.Value)
							if param.Key != "url" && param.Key != "access_token" {
								//log.Printf("Skipping key %s (2)", param.Key)
								continue
							}

							newParams = append(newParams, WorkflowAppActionParameter{
								Name:  param.Key,
								Value: param.Value,
							})
						}
					}
				}

				for _, param := range action.Parameters {
					if param.Configuration {
						continue
					}

					newParams = append(newParams, param)
				}
			} else {
				// Rebuild params with the right data. This is to prevent issues on the frontend
				for _, param := range action.Parameters {

					for _, authparam := range curAuth.Fields {
						if param.Name == authparam.Key {
							param.Value = authparam.Value
							//log.Printf("Name: %s - value: %s", param.Name, param.Value)
							//log.Printf("Name: %s - value: %s\n", param.Name, param.Value)
							break
						}
					}

					newParams = append(newParams, param)
				}
			}

			action.Parameters = newParams
		}

		action.LargeImage = ""
		if len(action.Label) == 0 {
			action.Label = action.ID
		}
		//log.Printf("LABEL: %s", action.Label)
		newActions = append(newActions, action)

		// If the node is NOT found, it's supposed to be set to SKIPPED,
		// as it's not a childnode of the startnode
		// This is a configuration item for the workflow itself.
		if len(workflowExecution.Results) > 0 {
			extra := 0
			for _, trigger := range workflowExecution.Workflow.Triggers {
				//log.Printf("Appname trigger (0): %s", trigger.AppName)
				if trigger.AppName == "User Input" || trigger.AppName == "Shuffle Workflow" {
					extra += 1
				}
			}

			/*
				defaultResults = []ActionResult{}
				for _, result := range workflowExecution.Results {
					if result.Status == "WAITING" {
						result.Status = "SUCCESS"
						result.Result = `{"success": true, "reason": "Continuing from user input"}`

						//log.Printf("Actions + extra = %d. Results = %d", len(workflowExecution.Workflow.Actions)+extra, len(workflowExecution.Results))
						if len(workflowExecution.Results) >= len(workflowExecution.Workflow.Actions)+extra {
							workflowExecution.Status = "FINISHED"
						} else {
							workflowExecution.Status = "EXECUTING"
						}
					}

					defaultResults = append(defaultResults, result)
				}
			*/
		} else if len(workflowExecution.Results) == 0 && !workflowExecution.Workflow.Configuration.StartFromTop {
			found := false
			for _, nodeId := range childNodes {
				if nodeId == action.ID {
					//log.Printf("Found %s", action.ID)
					found = true
				}
			}

			if !found {
				if action.ID == workflowExecution.Start {
					continue
				}

				//log.Printf("[WARNING] Set %s to SKIPPED as it's NOT a childnode of the startnode.", action.ID)
				curaction := Action{
					AppName:    action.AppName,
					AppVersion: action.AppVersion,
					Label:      action.Label,
					Name:       action.Name,
					ID:         action.ID,
				}
				//action
				//curaction.Parameters = []
				defaultResults = append(defaultResults, ActionResult{
					Action:        curaction,
					ExecutionId:   workflowExecution.ExecutionId,
					Authorization: workflowExecution.Authorization,
					Result:        `{"success": false, "reason": "Skipped because it's not under the startnode (1)"}`,
					StartedAt:     0,
					CompletedAt:   0,
					Status:        "SKIPPED",
				})
			}
		}
	}

	// Added fixes for e.g. URL's ending in /
	fixes := []string{"url"}
	for actionIndex, action := range workflowExecution.Workflow.Actions {
		if strings.ToLower(action.AppName) == "http" {
			continue
		}

		for paramIndex, param := range action.Parameters {
			if !param.Configuration {
				continue
			}

			if ArrayContains(fixes, strings.ToLower(param.Name)) {
				if strings.HasSuffix(param.Value, "/") {
					workflowExecution.Workflow.Actions[actionIndex].Parameters[paramIndex].Value = param.Value[0 : len(param.Value)-1]
				}
			}
		}
	}

	// Not necessary with comments at all
	workflowExecution.Workflow.Comments = []Comment{}
	for _, trigger := range workflowExecution.Workflow.Triggers {
		//log.Printf("[INFO] ID: %s vs %s", trigger.ID, workflowExecution.Start)
		if trigger.ID == workflowExecution.Start {
			if trigger.AppName == "User Input" {
				startFound = true
				break
			}
		}

		if trigger.AppName == "User Input" || trigger.AppName == "Shuffle Workflow" {
			found := false
			for _, node := range childNodes {
				if node == trigger.ID {
					found = true
					break
				}
			}

			if !found {
				//log.Printf("SHOULD SET TRIGGER %s TO BE SKIPPED", trigger.ID)

				curaction := Action{
					AppName:    "shuffle-subflow",
					AppVersion: trigger.AppVersion,
					Label:      trigger.Label,
					Name:       trigger.Name,
					ID:         trigger.ID,
				}

				found := false
				for _, res := range defaultResults {
					if res.Action.ID == trigger.ID {
						found = true
						break
					}
				}

				if !found {
					defaultResults = append(defaultResults, ActionResult{
						Action:        curaction,
						ExecutionId:   workflowExecution.ExecutionId,
						Authorization: workflowExecution.Authorization,
						Result:        `{"success": false, "reason": "Skipped because it's not under the startnode (2)"}`,
						StartedAt:     0,
						CompletedAt:   0,
						Status:        "SKIPPED",
					})
				}
			} else {
				// Replaces trigger with the subflow

			}
		}
	}

	if !startFound {
		if len(workflowExecution.Start) == 0 && len(workflowExecution.Workflow.Start) > 0 {
			workflowExecution.Start = workflow.Start
		} else if len(workflowExecution.Workflow.Actions) > 0 {
			workflowExecution.Start = workflowExecution.Workflow.Actions[0].ID
		} else {
			log.Printf("[ERROR] Startnode %s doesn't exist!!", workflowExecution.Start)
			return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Workflow action %s doesn't exist in workflow", workflowExecution.Start), errors.New(fmt.Sprintf(`Workflow start node "%s" doesn't exist. Exiting!`, workflowExecution.Start))
		}
	}

	//log.Printf("EXECUTION START: %s", workflowExecution.Start)

	// Verification for execution environments
	workflowExecution.Results = defaultResults
	workflowExecution.Workflow.Actions = newActions
	onpremExecution := true

	environments := []string{}
	if len(workflowExecution.ExecutionOrg) == 0 && len(workflow.ExecutingOrg.Id) > 0 {
		workflowExecution.ExecutionOrg = workflow.ExecutingOrg.Id
	}

	var allEnvs []Environment
	if len(workflowExecution.ExecutionOrg) > 0 {
		//log.Printf("[INFO] Executing ORG: %s", workflowExecution.ExecutionOrg)

		allEnvironments, err := GetEnvironments(ctx, workflowExecution.ExecutionOrg)
		if err != nil {
			log.Printf("Failed finding environments: %s", err)
			return WorkflowExecution{}, ExecInfo{}, fmt.Sprintf("Workflow environments not found for this org"), errors.New(fmt.Sprintf("Workflow environments not found for this org"))
		}

		for _, curenv := range allEnvironments {
			if curenv.Archived {
				continue
			}

			allEnvs = append(allEnvs, curenv)
		}
	} else {
		log.Printf("[ERROR] No org identified for execution of %s. Returning", workflowExecution.Workflow.ID)
		return WorkflowExecution{}, ExecInfo{}, "No org identified for execution", errors.New("No org identified for execution")
	}

	if len(allEnvs) == 0 {
		log.Printf("[ERROR] No active environments found for org: %s", workflowExecution.ExecutionOrg)
		return WorkflowExecution{}, ExecInfo{}, "No active environments found", errors.New(fmt.Sprintf("No active env found for org %s", workflowExecution.ExecutionOrg))
	}

	// Check if the actions are children of the startnode?
	imageNames := []string{}
	cloudExec := false
	for _, action := range workflowExecution.Workflow.Actions {

		// Verify if the action environment exists and append
		found := false
		for _, env := range allEnvs {
			if env.Name == action.Environment {
				found = true

				if env.Type == "cloud" {
					cloudExec = true
				} else if env.Type == "onprem" {
					onpremExecution = true
				} else {
					log.Printf("[ERROR] No handler for environment type %s", env.Type)
					return WorkflowExecution{}, ExecInfo{}, "No active environments found", errors.New(fmt.Sprintf("No handler for environment type %s", env.Type))
				}
				break
			}
		}

		if !found {
			if strings.ToLower(action.Environment) == "cloud" && project.Environment == "cloud" {
				log.Printf("[DEBUG] Couldn't find environment %s in cloud for some reason.", action.Environment)
			} else {
				log.Printf("[WARNING] Couldn't find environment %s. Maybe it's inactive?", action.Environment)
				return WorkflowExecution{}, ExecInfo{}, "Couldn't find the environment", errors.New(fmt.Sprintf("Couldn't find env %s in org %s", action.Environment, workflowExecution.ExecutionOrg))
			}
		}

		found = false
		for _, env := range environments {
			if env == action.Environment {
				found = true
				break
			}
		}

		// Check if the app exists?
		newName := action.AppName
		newName = strings.Replace(newName, " ", "-", -1)
		imageNames = append(imageNames, fmt.Sprintf("%s:%s_%s", baseDockerName, newName, action.AppVersion))

		if !found {
			environments = append(environments, action.Environment)
		}
	}

	//b, err := json.Marshal(workflowExecution)
	//if err == nil {
	//	log.Printf("LEN: %d", len(string(b)))
	//	//workflowExecution.ExecutionOrg.SyncFeatures = Org{}
	//}

	workflowExecution.Workflow.ExecutingOrg = OrgMini{
		Id: workflowExecution.Workflow.ExecutingOrg.Id,
	}
	workflowExecution.Workflow.Org = []OrgMini{
		workflowExecution.Workflow.ExecutingOrg,
	}

	// Means executing a subflow is happening
	if len(workflowExecution.ExecutionParent) > 0 {
		IncrementCache(ctx, workflowExecution.ExecutionOrg, "subflow_executions")
	}

	// NEW check for subflow
	// This is also handling triggers -> action translation now for subflow
	extra := 0
	newTriggers := []Trigger{}
	for _, trigger := range workflowExecution.Workflow.Triggers {
		if trigger.TriggerType != "SUBFLOW" && trigger.TriggerType != "USERINPUT" {
			newTriggers = append(newTriggers, trigger)
			continue
		}

		if trigger.TriggerType == "SUBFLOW" {
			//log.Printf("[INFO] Subflow trigger found during execution! envs: %#v", environments)

			// Find branch that has the subflow as destinationID
			foundenv := ""
			for _, branch := range workflowExecution.Workflow.Branches {
				if branch.DestinationID == trigger.ID {

					// FIX: May not work for subflow -> subflow if they are added in opposite order or something weird
					for _, action := range workflowExecution.Workflow.Actions {
						if action.ID == branch.SourceID {
							foundenv = action.Environment
							break
						}
					}

					if len(foundenv) > 0 {
						break
					}
				}
			}

			// Backup env :>
			if len(foundenv) == 0 && len(environments) > 0 {
				log.Printf("[ERROR] Fallback to environment %s for subflow (default)", environments[0])
				foundenv = environments[0]
			}

			// Setting to default?
			// environments := []string{}
			action := GetAction(workflowExecution, trigger.ID, foundenv)

			action.Label = trigger.Label
			action.ID = trigger.ID
			action.Name = "run_subflow"
			action.AppName = "shuffle-subflow"
			action.AppVersion = "1.0.0"

			action.Parameters = []WorkflowAppActionParameter{}
			for _, parameter := range trigger.Parameters {
				parameter.Variant = "STATIC_VALUE"
				if parameter.Name == "user_apikey" {
					continue
				}

				action.Parameters = append(action.Parameters, parameter)
			}

			action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
				Name:  "source_workflow",
				Value: workflowExecution.Workflow.ID,
			})

			action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
				Name:  "source_execution",
				Value: workflowExecution.ExecutionId,
			})

			action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
				Name:  "source_auth",
				Value: workflowExecution.Authorization,
			})

			action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
				Name:  "user_apikey",
				Value: workflowExecution.Authorization,
			})

			action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
				Name:  "source_node",
				Value: action.ID,
			})

			backendUrl := os.Getenv("BASE_URL")
			if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
				backendUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
			}

			if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 && strings.Contains(os.Getenv("SHUFFLE_CLOUDRUN_URL"), "http") {
				backendUrl = os.Getenv("SHUFFLE_CLOUDRUN_URL")
			}

			if len(backendUrl) > 0 {
				action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
					Name:  "backend_url",
					Value: backendUrl,
				})
			} else {
				log.Printf("[ERROR] No Backend URL found for subflow. May fail to connect properly.")

			}

			workflowExecution.Workflow.Actions = append(workflowExecution.Workflow.Actions, action)
		} else {
			newTriggers = append(newTriggers, trigger)
			extra += 1
		}
	}

	workflowExecution.Workflow.Triggers = newTriggers

	finished := ValidateFinished(ctx, extra, workflowExecution)
	if finished {
		log.Printf("[INFO][%s] Workflow already finished during startup. Is this correct?", workflowExecution.ExecutionId)
	}

	DeleteCache(ctx, fmt.Sprintf("workflowexecution_%s", workflowExecution.WorkflowId))
	DeleteCache(ctx, fmt.Sprintf("workflowexecution_%s_50", workflowExecution.WorkflowId))
	DeleteCache(ctx, fmt.Sprintf("workflowexecution_%s_100", workflowExecution.WorkflowId))

	workflowExecution.WorkflowId = workflowExecution.Workflow.ID

	return workflowExecution, ExecInfo{OnpremExecution: onpremExecution, Environments: environments, CloudExec: cloudExec, ImageNames: imageNames}, "", nil
}

func HealthCheckHandler(resp http.ResponseWriter, request *http.Request) {
	ret, err := project.Es.Info()
	if err != nil {
		log.Printf("[ERROR] Failed connecting to ES: %s", err)
		resp.WriteHeader(ret.StatusCode)
		resp.Write([]byte("Bad response from ES (1). Check logs for more details."))
		return
	}

	if ret.StatusCode >= 300 {
		resp.WriteHeader(ret.StatusCode)
		resp.Write([]byte(fmt.Sprintf("Bad response from ES - Status code %d", ret.StatusCode)))
		return
	}

	fmt.Fprint(resp, "OK")
}

// Extra validation sample to be used for workflow executions based on parent workflow instead of users' auth

// Check if the execution data has correct info in it! Happens based on subflows.
// 1. Parent workflow contains this workflow ID in the source trigger?
// 2. Parent workflow's owner is same org?
// 3. Parent execution auth is correct
func RunExecuteAccessValidation(request *http.Request, workflow *Workflow) (bool, string) {
	log.Printf("[DEBUG] Inside execute validation for workflow %s (%s)! Request method: %s", workflow.Name, workflow.ID, request.Method)

	//if request.Method == "POST" {
	ctx := GetContext(request)
	workflowExecution := &WorkflowExecution{}
	sourceExecution, sourceExecutionOk := request.URL.Query()["source_execution"]
	if !sourceExecutionOk {
		sourceExecution, sourceExecutionOk = request.URL.Query()["reference_execution"]
		if !sourceExecutionOk {
			log.Printf("[AUDIT] No source_execution or reference_execution in test validation")
			return false, ""
		}
	}

	if sourceExecutionOk && len(sourceExecution) > 0 {
		//log.Printf("[DEBUG] Got source exec %s", sourceExecution)
		newExec, err := GetWorkflowExecution(ctx, sourceExecution[0])
		if err != nil {
			log.Printf("[INFO] Failed getting source_execution in test validation based on %s", sourceExecution[0])
			return false, ""
		} else {
			workflowExecution = newExec
		}
	}

	if workflowExecution.ExecutionId == "" {
		log.Printf("[WARNING] No execution ID found. Bad auth.")
		return false, ""
	}

	sourceAuth, sourceAuthOk := request.URL.Query()["source_auth"]
	if !sourceAuthOk {
		sourceAuth, sourceAuthOk = request.URL.Query()["authorization"]
		if !sourceAuthOk {
			log.Printf("[AUDIT] No source auth found. Bad auth.")
			return false, ""
		}
	}

	if sourceAuthOk {
		//log.Printf("[DEBUG] Got source auth %s", sourceAuth)
		// Must origin from a parent workflow")

		if sourceAuth[0] != workflowExecution.Authorization {
			log.Printf("[AUDIT] Bad authorization for workflowexecution defined.")
			return false, ""
		}

		// Check if workflow is in waiting stage
		// If it is, accept from this point already, as it's a user input action
		if workflowExecution.Status == "WAITING" {
			return true, ""
		}

	}

	// When reaching here, authentication is done, but not authorization.
	// Need to verify the workflow, and whether it SHOULD have access to execute it.
	sourceWorkflow, sourceWorkflowOk := request.URL.Query()["source_workflow"]
	if sourceWorkflowOk {
		//log.Printf("[DEBUG] Got source workflow %s", sourceWorkflow)
		_ = sourceWorkflow

		// Source workflow = parent
		// This workflow = child

		//if sourceWorkflow[0] != workflow.ID {
		//	log.Printf("[DEBUG] Bad workflow in execution.")
		//	return false, ""
		//}

	} else {
		log.Printf("[AUDIT] Did NOT get source workflow in subflow execution")
		return false, ""
	}

	if workflow.OrgId != workflowExecution.Workflow.OrgId || workflow.ExecutingOrg.Id != workflowExecution.Workflow.ExecutingOrg.Id || workflow.OrgId == "" {
		log.Printf("[AUDIT] Bad org ID in workflowexecution defined.")
		return false, ""
	}

	// 1. Parent workflow contains this workflow ID in the source trigger?
	// 2. Parent workflow's owner is same org?
	// 3. Parent execution auth is correct

	sourceNode, sourceNodeOk := request.URL.Query()["source_node"]
	if !sourceNodeOk {
		log.Printf("[AUDIT] Couldn't find source node that started the execution")
		return false, ""
	}

	// SHOULD be executed by a trigger in the parent.
	for _, trigger := range workflowExecution.Workflow.Triggers {
		if sourceNode[0] == trigger.ID {
			return true, workflowExecution.ExecutionOrg
		}
	}

	for _, action := range workflowExecution.Workflow.Actions {
		if sourceNode[0] == action.ID {
			return true, workflowExecution.ExecutionOrg
		}
	}
	//}

	log.Printf("[AUDIT] Bad auth for workflowexecution (bottom).")

	return false, ""
}

// Significantly slowed down everything. Just returning for now.
func findReferenceAppDocs(ctx context.Context, allApps []WorkflowApp) []WorkflowApp {
	newApps := []WorkflowApp{}

	// Skipping this for now as it makes things slow
	return allApps

	for _, app := range allApps {
		if len(app.ReferenceInfo.DocumentationUrl) > 0 && strings.HasPrefix(app.ReferenceInfo.DocumentationUrl, "https://raw.githubusercontent.com/Shuffle") && strings.Contains(app.ReferenceInfo.DocumentationUrl, ".md") {
			// Should find documentation from the github (only if github?) and add it to app.Documentation before caching
			//log.Printf("DOCS: %s", app.ReferenceInfo.DocumentationUrl)
			documentationData, err := DownloadFromUrl(ctx, app.ReferenceInfo.DocumentationUrl)
			if err != nil {
				log.Printf("[ERROR] Failed getting data: %s", err)
			} else {
				app.Documentation = string(documentationData)
			}
		}

		//if app.Documentation == "" && strings.ToLower(app.Name) == "discord" {
		if app.Documentation == "" {
			referenceUrl := ""

			if app.Generated {
				//log.Printf("[DEBUG] Should look in the OpenAPI folder")
				baseUrl := "https://raw.githubusercontent.com/Shuffle/openapi-apps/master/docs"

				newName := strings.ToLower(strings.Replace(strings.Replace(app.Name, " ", "_", -1), "-", "_", -1))
				referenceUrl = fmt.Sprintf("%s/%s.md", baseUrl, newName)
			} else {
				//log.Printf("[DEBUG] Should look in the Python-apps folder")
			}

			if len(referenceUrl) > 0 {
				//log.Printf("REF: %s", referenceUrl)

				documentationData, err := DownloadFromUrl(ctx, referenceUrl)
				if err != nil {
					log.Printf("[ERROR] Failed getting documentation data for app %s: %s", app.Name, err)
				} else {
					//log.Printf("[INFO] Added documentation from github for %s", app.Name)
					app.Documentation = string(documentationData)
				}
			}
		}

		newApps = append(newApps, app)
	}

	return newApps
}

func EchoOpenapiData(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Just here to verify that the user is logged in
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[DEBUG] Api authentication failed in download Yaml echo: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to echo OpenAPI data: %s (%s)", user.Username, user.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Bodyreader err: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed reading body"}`))
		return
	}

	newbody := string(body)
	newbody = strings.TrimSpace(newbody)
	if strings.HasPrefix(newbody, "\"") {
		newbody = newbody[1:len(newbody)]
	}

	if strings.HasSuffix(newbody, "\"") {
		newbody = newbody[0 : len(newbody)-1]
	}

	// Rewrite to download proper one from Github even without raw path
	if strings.Contains(newbody, "https://github.com/") {
		// https://github.com/AdguardTeam/AdGuardHome/blob/master/openapi/openapi.yaml
		// https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/openapi/openapi.yaml
		// https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/openapi/openapi.yaml

		urlsplit := strings.Split(newbody, "/")
		if len(urlsplit) > 6 {
			log.Printf("[DEBUG] Rewriting github URL %s.", newbody)
			ghuser := urlsplit[3]
			repo := urlsplit[4]
			branch := urlsplit[6]
			path := strings.Join(urlsplit[7:len(urlsplit)], "/")
			newbody = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", ghuser, repo, branch, path)
		}
	}

	log.Printf("[DEBUG] Downloading content from %s", newbody)

	req, err := http.NewRequest("GET", newbody, nil)
	if err != nil {
		log.Printf("[ERROR] Requestbuilder err: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Failed building request"}`))
		return
	}

	httpClient := &http.Client{}
	newresp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] Grabbing error: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed making remote request to get the data"}`)))
		return
	}
	defer newresp.Body.Close()

	urlbody, err := ioutil.ReadAll(newresp.Body)
	if err != nil {
		log.Printf("[ERROR] URLbody error: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Can't get data from selected uri"}`)))
		return
	}

	if newresp.StatusCode >= 400 {
		resp.WriteHeader(201)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "%s"}`, urlbody)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(urlbody)
}

func GetFrameworkConfiguration(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Just here to verify that the user is logged in
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[DEBUG] Api authentication failed in get detection framework: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
		return
	}

	ctx := GetContext(request)
	org, err := GetOrg(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Error getting org: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	//log.Printf("Framework: %s", org.SecurityFramework)
	newjson, err := json.Marshal(org.SecurityFramework)
	if err != nil {
		log.Printf("[ERROR] Failed marshal in get security framework: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking framework. Contact us to get it fixed."}`)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func SetFrameworkConfiguration(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Just here to verify that the user is logged in
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[DEBUG] Api authentication failed in set detection framework: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Error with body read: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	type parsedValue struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		ID          string `json:"id"`
		LargeImage  string `json:"large_image"`
		Description string `json:"description"`
	}

	var value parsedValue
	err = json.Unmarshal(body, &value)
	if err != nil {
		log.Printf("[WARNING] Error with unmarshal tmpBody in frameworkconfig: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)
	org, err := GetOrg(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Error getting org in set framework: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	app := &WorkflowApp{
		Name:        "",
		Description: "",
		ID:          "",
		LargeImage:  "",
	}

	// System for replacing an app if it's not defined
	if value.ID != "remove" {
		app, err = GetApp(ctx, value.ID, user, false)
		if err != nil {

			if project.Environment == "cloud" {
				log.Printf("[ERROR] Error getting app %s in set framework: %s", value.ID, err)
				resp.WriteHeader(401)
				resp.Write([]byte(`{"success": false}`))
				return
			} else {
				// Forwarded from Algolia in the frontend
				app.Name = value.Name
				app.ID = value.Name
				app.Description = value.Description
				app.LargeImage = value.LargeImage
			}
		}

		if project.Environment == "cloud" && !app.Sharing && app.Public {
			log.Printf("[WARNING] Error setting app %s for org %s as it's not public.", value.ID, err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	if value.Type == "email" {
		value.Type = "comms"
	}

	if value.Type == "eradication" {
		value.Type = "edr"
	}

	// 1. Check if the app exists and the user has access to it. If public/sharing ->

	if strings.ToLower(value.Type) == "siem" {
		org.SecurityFramework.SIEM.Name = app.Name
		org.SecurityFramework.SIEM.Description = app.Description
		org.SecurityFramework.SIEM.ID = app.ID
		org.SecurityFramework.SIEM.LargeImage = app.LargeImage
	} else if strings.ToLower(value.Type) == "network" {
		org.SecurityFramework.Network.Name = app.Name
		org.SecurityFramework.Network.Description = app.Description
		org.SecurityFramework.Network.ID = app.ID
		org.SecurityFramework.Network.LargeImage = app.LargeImage
	} else if strings.ToLower(value.Type) == "edr" || strings.ToLower(value.Type) == "edr & av" || strings.ToLower(value.Type) == "eradication" {
		org.SecurityFramework.EDR.Name = app.Name
		org.SecurityFramework.EDR.Description = app.Description
		org.SecurityFramework.EDR.ID = app.ID
		org.SecurityFramework.EDR.LargeImage = app.LargeImage
	} else if strings.ToLower(value.Type) == "cases" {
		org.SecurityFramework.Cases.Name = app.Name
		org.SecurityFramework.Cases.Description = app.Description
		org.SecurityFramework.Cases.ID = app.ID
		org.SecurityFramework.Cases.LargeImage = app.LargeImage
	} else if strings.ToLower(value.Type) == "iam" {
		org.SecurityFramework.IAM.Name = app.Name
		org.SecurityFramework.IAM.Description = app.Description
		org.SecurityFramework.IAM.ID = app.ID
		org.SecurityFramework.IAM.LargeImage = app.LargeImage
	} else if strings.ToLower(value.Type) == "assets" {
		org.SecurityFramework.Assets.Name = app.Name
		org.SecurityFramework.Assets.Description = app.Description
		org.SecurityFramework.Assets.ID = app.ID
		org.SecurityFramework.Assets.LargeImage = app.LargeImage
	} else if strings.ToLower(value.Type) == "intel" {
		org.SecurityFramework.Intel.Name = app.Name
		org.SecurityFramework.Intel.Description = app.Description
		org.SecurityFramework.Intel.ID = app.ID
		org.SecurityFramework.Intel.LargeImage = app.LargeImage
	} else if strings.ToLower(value.Type) == "comms" || strings.ToLower(value.Type) == "communication" || strings.ToLower(value.Type) == "email" {
		org.SecurityFramework.Communication.Name = app.Name
		org.SecurityFramework.Communication.Description = app.Description
		org.SecurityFramework.Communication.ID = app.ID
		org.SecurityFramework.Communication.LargeImage = app.LargeImage
	} else {
		log.Printf("[WARNING] No handler for type %s in app framework during update of app %s", value.Type, app.Name)
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Counting up for the getting started piece
	cnt := 0
	if len(org.SecurityFramework.SIEM.Name) > 0 {
		cnt += 1
	}

	if len(org.SecurityFramework.Intel.Name) > 0 {
		cnt += 1
	}

	if len(org.SecurityFramework.Communication.Name) > 0 {
		cnt += 1
	}

	if len(org.SecurityFramework.Assets.Name) > 0 {
		cnt += 1
	}

	if len(org.SecurityFramework.IAM.Name) > 0 {
		cnt += 1
	}

	if len(org.SecurityFramework.Cases.Name) > 0 {
		cnt += 1
	}

	if len(org.SecurityFramework.EDR.Name) > 0 {
		cnt += 1
	}

	if len(org.SecurityFramework.Network.Name) > 0 {
		cnt += 1
	}

	// Add app as active for org too
	if len(app.ID) > 0 && !ArrayContains(org.ActiveApps, app.ID) {
		org.ActiveApps = append(org.ActiveApps, app.ID)
	}

	for tutorialIndex, tutorial := range org.Tutorials {
		if tutorial.Name == "Find relevant apps" {
			org.Tutorials[tutorialIndex].Description = fmt.Sprintf("%d out of %d apps configured. Find more relevant apps in the search bar.", cnt, 8)

			if cnt > 0 {
				org.Tutorials[tutorialIndex].Done = true
			}
		}
	}

	err = SetOrg(ctx, *org, org.Id)
	if err != nil {
		log.Printf("[WARNING] Failed setting app framework for org %s: %s", org.Name, err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed updating organization info. Please contact us if this persists."}`))
		return
	} else {
		DeleteCache(ctx, fmt.Sprintf("apps_%s", user.Id))
		DeleteCache(ctx, fmt.Sprintf("apps_%s", user.ActiveOrg.Id))
		DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-100"))
		DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-500"))
		DeleteCache(ctx, fmt.Sprintf("workflowapps-sorted-1000"))
		DeleteCache(ctx, "all_apps")
		DeleteCache(ctx, fmt.Sprintf("user_%s", user.Username))
		DeleteCache(ctx, fmt.Sprintf("user_%s", user.Id))
	}

	if value.ID != "remove" {
		log.Printf("[DEBUG] Successfully updated app framework type %s to app %s (%s) for org %s (%s)!", value.Type, app.Name, app.ID, org.Name, org.Id)
	} else {
		log.Printf("[DEBUG] Successfully REMOVED app framework type %s for org %s (%s)!", value.Type, org.Name, org.Id)
	}

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

type flushWriter struct {
	f http.Flusher
	w io.Writer
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}
	return
}

func HandleStreamWorkflow(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	//// Removed check here as it may be a public workflow
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[AUDIT] Api authentication failed in getting specific workflow (stream): %s. Continuing because it may be public.", err)
	}

	location := strings.Split(request.URL.String(), "/")

	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if strings.Contains(fileId, "?") {
		fileId = strings.Split(fileId, "?")[0]
	}

	if len(fileId) != 36 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow ID when getting workflow is not valid"}`))
		return
	}

	//ctx := GetContext(request)
	ctx := GetContext(request)
	workflow, err := GetWorkflow(ctx, fileId)
	if err != nil {
		log.Printf("[WARNING] Workflow %s doesn't exist.", fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed finding workflow."}`))
		return
	}

	if user.Id != workflow.Owner || len(user.Id) == 0 {
		if workflow.OrgId == user.ActiveOrg.Id && (user.Role == "admin" || user.Role == "org-reader") {
			log.Printf("[AUDIT] User %s is accessing workflow %s as admin (stream edit workflow)", user.Username, workflow.ID)
		} else if workflow.Public {
			log.Printf("[AUDIT] Letting user %s access workflow %s for streaming because it's public", user.Username, workflow.ID)
		} else if project.Environment == "cloud" && user.Verified == true && user.Active == true && user.SupportAccess == true && strings.HasSuffix(user.Username, "@shuffler.io") {
			log.Printf("[AUDIT] Letting verified support admin %s access workflow %s", user.Username, workflow.ID)
		} else {
			log.Printf("[AUDIT] Wrong user (%s) for workflow %s (get workflow)", user.Username, workflow.ID)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	resp.Header().Set("Connection", "Keep-Alive")
	resp.Header().Set("X-Content-Type-Options", "nosniff")

	conn, ok := resp.(http.Flusher)
	if !ok {
		log.Printf("[ERROR] Flusher error: %s", ok)
		http.Error(resp, "Streaming supported on AppEngine", http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "text/event-stream")
	resp.WriteHeader(http.StatusOK)

	sessionKey := fmt.Sprintf("%s_stream", workflow.ID)
	previousCache := []byte{}
	for {
		cache, err := GetCache(ctx, sessionKey)
		if err == nil {
			cacheData := []byte(cache.([]uint8))
			if string(previousCache) == string(cacheData) {
				//log.Printf("[DEBUG] Still same cache for %s", user.Id)
			} else {
				// A way to only check for data from other people
				if (len(user.Id) > 0 && !strings.Contains(string(cacheData), user.Id)) || len(user.Id) == 0 {
					//fw.Write(cacheData)
					//w.Write(cacheData)

					_, err := fmt.Fprintf(resp, string(cacheData))
					if err != nil {
						log.Printf("[ERROR] Failed in writing stream to user: %s", err)
					} else {
						previousCache = cacheData
						conn.Flush()
					}

				} else {
					//log.Printf("[ERROR] NEW cache for %s (2) - NOT sending: %s.", user.Id, cacheData)

					previousCache = cacheData
				}

			}
		} else {
			//log.Printf("[DEBUG] Failed getting cache for %s: %s", user.Id, err)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func HandleStreamWorkflowUpdate(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	//// Removed check here as it may be a public workflow
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[AUDIT] Api authentication failed in getting specific workflow (stream update): %s. Continuing because it may be public.", err)
	}

	location := strings.Split(request.URL.String(), "/")

	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if strings.Contains(fileId, "?") {
		fileId = strings.Split(fileId, "?")[0]
	}

	if len(fileId) != 36 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Workflow ID when getting workflow is not valid"}`))
		return
	}

	ctx := GetContext(request)
	workflow, err := GetWorkflow(ctx, fileId)
	if err != nil {
		log.Printf("[WARNING] Workflow %s doesn't exist.", fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed finding workflow."}`))
		return
	}

	if user.Id != workflow.Owner || len(user.Id) == 0 {
		if workflow.OrgId == user.ActiveOrg.Id && (user.Role == "admin" || user.Role == "org-reader") {
			log.Printf("[AUDIT] User %s is accessing workflow %s as admin (get workflow)", user.Username, workflow.ID)

		} else if project.Environment == "cloud" && user.Verified == true && user.SupportAccess == true && user.Role == "admin" {
			log.Printf("[AUDIT] Letting verified support admin %s access workflow %s", user.Username, workflow.ID)

		} else {
			log.Printf("[AUDIT] Wrong user (%s) for workflow %s (get workflow stream)", user.Username, workflow.ID)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Error with body read in workflow stream: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Literally just dumping them in, as they're supposed to be overwritten continuously
	// PS: This is NOT an ideal process, and broadcasting should be handled differently
	//log.Printf("Body to update: %s", string(body))
	sessionKey := fmt.Sprintf("%s_stream", workflow.ID)
	err = SetCache(ctx, sessionKey, body, 30)
	if err != nil {
		log.Printf("[WARNING] Failed setting cache for apikey: %s", err)
	}

	resp.WriteHeader(200)
	resp.Write([]byte("OK"))
}

func LoadUsecases(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get usecases. Continuing anyway: %s", err)
		//resp.WriteHeader(401)
		//resp.Write([]byte(`{"success": false}`))
		//return
	}

	// FIXME: Load for the specific org and have structs for it all
	_ = user

	//ctx := GetContext(request)

	resp.WriteHeader(200)
	resp.Write([]byte(GetUsecaseData()))
}

func UpdateUsecases(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get usecases. Continuing anyway: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	// Needs to be an active shuffler.io account to update
	if project.Environment == "cloud" && !strings.HasSuffix(user.Username, "@shuffler.io") {
		resp.WriteHeader(403)
		resp.Write([]byte(`{"success": false, "reason": "Can't change framework info"}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Error with body read for usecase update: %s", err)
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var usecase Usecase
	err = json.Unmarshal(body, &usecase)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshaling usecase: %s", err)
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	usecase.Success = true
	usecase.Name = strings.Replace(usecase.Name, " ", "_", -1)
	usecase.Name = url.QueryEscape(usecase.Name)
	log.Printf("[DEBUG] Updated usecase %s as user %s (%s)", usecase.Name, user.Username, user.Id)
	usecase.EditedBy = user.Id
	ctx := GetContext(request)
	err = SetUsecase(ctx, usecase)
	if err != nil {
		log.Printf("[ERROR] Failed updating usecase: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

func HandleGetUsecase(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	_, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get usecase (1). Continuing anyway: %s", err)
		//resp.WriteHeader(401)
		//resp.Write([]byte(`{"success": false}`))
		//return
	}

	var name string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 5 {
			log.Printf("[ERROR] Path too short: %d", len(location))
			resp.WriteHeader(400)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		name = location[5]
	}

	ctx := GetContext(request)
	usecase, err := GetUsecase(ctx, name)
	if err != nil {
		log.Printf("[ERROR] Failed getting usecase %s: %s", name, err)
		usecase.Success = true
		usecase.Name = name
		//resp.WriteHeader(400)
		//resp.Write([]byte(`{"success": false}`))
		//return
	} else {
		usecase.Success = true
	}

	// Hardcoding until we have something good for open source + cloud
	replacedName := strings.Replace(strings.ToLower(usecase.Name), " ", "_", -1)
	if replacedName == "email_management" {
		usecase.ExtraButtons = []ExtraButton{
			ExtraButton{
				Name:  "IMAP",
				App:   "Email",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/email_ec25da1fdbf18934ca468788b73bec32.png",
				Link:  "https://shuffler.io/workflows/b65d180c-4d27-4cb6-8128-3687a08aadb3",
				Type:  "communication",
			},
			ExtraButton{
				Name:  "Gmail",
				App:   "Gmail",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/Gmail_794e51c3c1a8b24b89ccc573a3defc47.png",
				Link:  "https://shuffler.io/workflows/e506060f-0c58-4f95-a0b8-f671103d78e5",
				Type:  "communication",
			},
			ExtraButton{
				Name:  "Outlook",
				App:   "Outlook Graph",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/Outlook_graph_d71641a57deeee8149df99080adebeb7.png",
				Link:  "https://shuffler.io/workflows/3862ed8f-7801-4393-8524-05de8f8a401d",
				Type:  "communication",
			},
		}
	} else if replacedName == "edr_to_ticket" {
		usecase.ExtraButtons = []ExtraButton{
			ExtraButton{
				Name:  "Velociraptor",
				App:   "Velociraptor",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/velociraptor_63de9fc91bcb4813d9c58cc6efd49b33.png",
				Link:  "https://shuffler.io/apps/63de9fc91bcb4813d9c58cc6efd49b33",
				Type:  "edr",
			},
			ExtraButton{
				Name:  "Carbon Black",
				App:   "Carbon Black",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/Carbon_Black_Response_e9fa2602ea6baafffa4b5eec722095d3.png",
				Link:  "https://shuffler.io/apps/e9fa2602ea6baafffa4b5eec722095d3",
				Type:  "edr",
			},
			ExtraButton{
				Name:  "Crowdstrike",
				App:   "Crowdstrike",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/Crowdstrike_Falcon_7a66ce3c26e0d724f31f1ebc9a7a41b4.png",
				Link:  "https://shuffler.io/apps/7a66ce3c26e0d724f31f1ebc9a7a41b4",
				Type:  "edr",
			},
		}
	} else if replacedName == "siem_to_ticket" {
		usecase.ExtraButtons = []ExtraButton{
			ExtraButton{
				Name:  "Wazuh",
				App:   "Wazuh",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/Wazuh_fb715a176a192687e95e9d162186c97f.png",
				Link:  "https://shuffler.io/workflows/bb45124c-d39e-4acc-a5d9-f8aa526042b5",
				Type:  "siem",
			},
			ExtraButton{
				Name:  "Splunk",
				App:   "Splunk",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/Splunk_Splunk_e352462c6d2f0a692281600d96002a45.png",
				Link:  "https://shuffler.io/apps/441a2d85f6c1e8408dd1ee1e804cd241",
				Type:  "siem",
			},
			ExtraButton{
				Name:  "QRadar",
				App:   "QRadar",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/QRadar_4fe358bd204f672d37c55b4f1d48ccdb.png",
				Link:  "https://shuffler.io/apps/96a3d95a2a73cfdb51ea4a394287ed33",
				Type:  "siem",
			},
		}
	} else if replacedName == "chatops" {
		usecase.ExtraButtons = []ExtraButton{
			ExtraButton{
				Name:  "Webex",
				App:   "Webex",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/Webex_1f6f2fc4fd399597e98ff34f78f56c45.png",
				Link:  "https://shuffler.io/workflows/88e16093-37b7-41cf-b02b-d1ca0e737993",
				Type:  "communication",
			},
			ExtraButton{
				Name:  "Teams",
				App:   "Microsoft Teams",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/Microsoft_Teams_User_Access_4826c529f8082205a4b926ac9f1dfcfb.png",
				Link:  "https://shuffler.io/apps/4826c529f8082205a4b926ac9f1dfcfb",
				Type:  "communication",
			},
			ExtraButton{
				Name:  "Slack",
				App:   "Slack",
				Image: "https://storage.googleapis.com/shuffle_public/app_images/Slack_Web_API_f63a65ddf0ee369845b6918575d47fc1.png",
				Link:  "https://shuffler.io/workflows/0a7eeca9-e056-40e5-9a70-f078937c6055",
				Type:  "communication",
			},
		}
	}

	newjson, err := json.Marshal(usecase)
	if err != nil {
		log.Printf("[ERROR] Failed marshal in get usecase: %s", err)
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking data"}`)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(newjson)
}

func GetBackendexecution(ctx context.Context, executionId, authorization string) (WorkflowExecution, error) {
	exec := WorkflowExecution{}

	// Is polling the backend actually correct?
	// Or should worker/backend talk to itself?
	backendUrl := os.Getenv("BASE_URL")
	if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
		backendUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
	}

	if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
		backendUrl = os.Getenv("SHUFFLE_CLOUDRUN_URL")
	}

	// Should this be without worker? :thinking:
	resultUrl := fmt.Sprintf("%s/api/v1/streams/results", backendUrl)

	topClient := &http.Client{
		Transport: &http.Transport{
			Proxy: nil,
		},
	}

	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")
	if len(httpProxy) > 0 || len(httpsProxy) > 0 {
		topClient = &http.Client{}
	} else {
		if len(httpProxy) > 0 {
			log.Printf("Running with HTTP proxy %s (env: HTTP_PROXY)", httpProxy)
		}
		if len(httpsProxy) > 0 {
			log.Printf("Running with HTTPS proxy %s (env: HTTPS_PROXY)", httpsProxy)
		}
	}

	requestData := ActionResult{
		ExecutionId:   executionId,
		Authorization: authorization,
	}

	data, err := json.Marshal(requestData)
	if err != nil {
		log.Printf("[WARNING] Failed parent init marshal: %s", err)
		return exec, err
	}

	req, err := http.NewRequest(
		"POST",
		resultUrl,
		bytes.NewBuffer([]byte(data)),
	)

	newresp, err := topClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed making subflow request (1): %s. Is URL valid: %s", err, resultUrl)
		return exec, err
	}

	body, err := ioutil.ReadAll(newresp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed reading parent body: %s", err)
		return exec, err
	}
	//log.Printf("BODY (%d): %s", newresp.StatusCode, string(body))

	if newresp.StatusCode != 200 {
		log.Printf("[ERROR] Bad statuscode getting execution (2) with URL %s: %d, %s", resultUrl, newresp.StatusCode, string(body))
		return exec, errors.New(fmt.Sprintf("Bad statuscode: %d", newresp.StatusCode))
	}

	err = json.Unmarshal(body, &exec)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshalling execution: %s", err)
		return exec, err
	}

	if exec.Status == "FINISHED" || exec.Status == "FAILURE" {
		cacheKey := fmt.Sprintf("workflowexecution_%s", executionId)
		err = SetCache(ctx, cacheKey, body, 30)
		if err != nil {
			log.Printf("[WARNING] Failed setting cache for workflowexec key %s: %s", cacheKey, err)
		}
	}

	return exec, nil
}

func AddPriority(org Org, priority Priority, updated bool) (*Org, bool) {
	found := false
	for _, p := range org.Priorities {
		if p.Name == priority.Name || (p.Type == priority.Type && p.Active) {
			found = true
			break
		}
	}

	if !found {
		priority.Active = true
		org.Priorities = append(org.Priorities, priority)
		updated = true
	}

	return &org, updated
}

// Watch academy - Shuffle 101 if you're new
// Check notifications
// Check if all apps have been discovered
// Check if notification workflow is made
// Check workflows vs usecases
// Check based on org configuration (name, environments, apps...)
// Workflows not finished / without a parent workflow/trigger
// Workflows not in production

// Should just be based on cache, not queries - keep it fast!
func GetPriorities(ctx context.Context, user User, org *Org) ([]Priority, error) {
	// Get usecases -> Check which aren't done based on priorities
	// 1. Check what apps are selected. Are email, edr and siem in there?
	// If not - autocorrect based on workflows' apps?
	//log.Printf("[DEBUG] SecurityFramework: %s", org.SecurityFramework)

	// First prio: Find these & attach usecases?
	// Only set these if cache is set for the user
	orgUpdated := false
	updated := false
	if project.CacheDb == false {
		// Not checking as cache is used for all checks
		return org.Priorities, nil
	}

	if len(org.Defaults.NotificationWorkflow) == 0 {
		org, updated = AddPriority(*org, Priority{
			Name:        fmt.Sprintf("You haven't defined a notification workflow yet."),
			Description: "Notification workflows are used to automate your notification handling. These can be used to alert yourself in other systems when issues are found in your current- or sub-organizations",
			Type:        "notifications",
			Active:      true,
			URL:         fmt.Sprintf("/admin?admin_tab=organization"),
			Severity:    2,
		}, updated)

		if updated {
			orgUpdated = true
		}
	}

	// Notify about hybrid
	if project.Environment == "cloud" {
		org, updated = AddPriority(*org, Priority{
			Name:        fmt.Sprintf("Try Hybrid Shuffle by connecting environments"),
			Description: "Hybrid Shuffle allows you to connect Shuffle to your local datacenter(s) internal resources, and get the results in the cloud.",
			Type:        "hybrid",
			Active:      true,
			URL:         fmt.Sprintf("/admin?tab=environments"),
			Severity:    3,
		}, updated)
	} else {
		org, updated = AddPriority(*org, Priority{
			Name:        fmt.Sprintf("Get more functionality by connecting to the cloud"),
			Description: "Get access to webhooks, schedules and other functions by connecting to the cloud. Try it out for free, or contact our team if you want to learn more!",
			Type:        "hybrid",
			Active:      true,
			URL:         fmt.Sprintf("/admin?admin_tab=cloud_sync"),
			Severity:    2,
		}, updated)
	}

	var notifications []Notification
	cache, err := GetCache(ctx, fmt.Sprintf("notifications_%s", org.Id))
	if err == nil {
		cacheData := []byte(cache.([]uint8))
		//log.Printf("CACHEDATA: %s", cacheData)
		err = json.Unmarshal(cacheData, &notifications)
		if err == nil && len(notifications) > 0 {
			org, updated = AddPriority(*org, Priority{
				Name:        fmt.Sprintf("You have %d unhandled notifications.", len(notifications)),
				Description: "Notifications help make your workflow infrastructure stable. Click the notification icon in the top right to see all open ones.",
				Type:        "notifications",
				Active:      true,
				URL:         fmt.Sprintf("/notifications"),
				Severity:    1,
			}, updated)

			if updated {
				orgUpdated = true
			}
		}
	} else {
		//log.Printf("[DEBUG] Failed getting cache for org: %s", err)
	}

	if len(org.MainPriority) == 0 {
		// Just choosing something for them, e.g. basic usecase building

		org.MainPriority = "1. Collect"
		orgUpdated = true
	}

	org, orgUpdated = GetWorkflowSuggestions(ctx, user, org, orgUpdated)

	if orgUpdated {
		log.Printf("[DEBUG] Should update org with %d priorities", len(org.Priorities))
		SetOrg(ctx, *org, org.Id)
	}

	return org.Priorities, nil
	//return []Priority{}, nil
}

// Sorts an org list in order to make ChildOrgs appear under their parent org
func SortOrgList(orgs []OrgMini) []OrgMini {
	// Creates parentorg map
	parentOrgs := map[string][]OrgMini{}
	for _, org := range orgs {
		if len(org.CreatorOrg) == 0 && len(org.ChildOrgs) > 0 {
			parentOrgs[org.Id] = []OrgMini{}
		} else if len(org.CreatorOrg) == 0 {
			// No childorgs, but isn't a parentorg either
			parentOrgs[org.Id] = []OrgMini{}
		} else {
			// Child orgs go here
		}
	}

	noParentOrg := []OrgMini{}
	for _, org := range orgs {
		// Check if parent in parentOrgs map
		if len(org.CreatorOrg) == 0 {
			continue
		}

		if val, ok := parentOrgs[org.CreatorOrg]; ok {
			parentOrgs[org.CreatorOrg] = append(val, org)
		} else {
			noParentOrg = append(noParentOrg, org)
		}
	}

	newOrgs := []OrgMini{}
	for key, value := range parentOrgs {
		// Find key in orgs
		found := false
		for _, org := range orgs {
			if org.Id == key {
				found = true
				newOrgs = append(newOrgs, org)
				break
			}
		}

		if found {
			for _, childorg := range value {
				newOrgs = append(newOrgs, childorg)
			}
		}
	}

	// Adding orgs where parentorg is unavailable
	// They should probably be under some "inactive" parentorg..
	newOrgs = append(newOrgs, noParentOrg...)

	return newOrgs
}

func findMissingChildren(ctx context.Context, workflowExecution *WorkflowExecution, children map[string][]string, inputNode string, checkedNodes []string) []string {
	nextActions := []string{}

	if ArrayContains(checkedNodes, inputNode) {
		return nextActions
	}

	checkedNodes = append(checkedNodes, inputNode)
	parentRan := false
	for _, result := range workflowExecution.Results {
		if result.Action.ID == inputNode {
			parentRan = true
		}
	}

	if !parentRan {
		//log.Printf("Should run parent first.")
		return []string{inputNode}
	} else {
		// Starting from the startnode, go through the workflow one level at a time
		foundCnt := 0
		for _, child := range children[inputNode] {
			// Check if the parent and its childs have a result
			found := false
			for _, result := range workflowExecution.Results {
				if result.Action.ID == child {
					foundCnt += 1
					found = true
					break
				}
			}

			if !found {
				// Due to being too fast cleared
				//cacheId := fmt.Sprintf("%s_%s", workflowExecution.ExecutionId, child)
				//_, err := GetCache(ctx, cacheId)
				//if err != nil {
				//	//log.Printf("[INFO] Missing execution (2): %s", child)
				//	nextActions = append(nextActions, child)
				//	continue
				//}

				cacheId := fmt.Sprintf("%s_%s_result", workflowExecution.ExecutionId, child)
				_, err := GetCache(ctx, cacheId)
				if err != nil {
					//log.Printf("[INFO] Missing execution (2): %s", child)
					nextActions = append(nextActions, child)
				}
			}
		}

		if foundCnt == len(children[inputNode]) {
			//log.Printf("All nodes done. Check their child results. Child nodes: %d, found: %d", len(children[inputNode]), foundCnt)
			// Run child nodes of this!
			nextActions = []string{}

			// Randomize order as to keep digging
			//for _, child := range rand.Perm(children[inputNode]) {
			for _, child := range children[inputNode] {
				next := findMissingChildren(ctx, workflowExecution, children, child, checkedNodes)

				// Only doing one are at a time as to SLOWLY dig down into it
				if len(next) > 0 {
					nextActions = append(nextActions, next...)
					break
				}
			}

			return nextActions
		} else {
			//log.Printf("Missing nodes. Found: %d, Expected: %d", foundCnt, len(children[inputNode]))
		}
	}

	return nextActions
}

// Finds next actions that aren't already executed and don't have results
func CheckNextActions(ctx context.Context, workflowExecution *WorkflowExecution) []string {
	extra := 0
	parents := map[string][]string{}
	children := map[string][]string{}
	nextActions := []string{}

	inputNode := workflowExecution.Start
	if len(workflowExecution.Results) == 0 {
		return []string{inputNode}
	}

	for _, trigger := range workflowExecution.Workflow.Triggers {
		if trigger.TriggerType != "SUBFLOW" && trigger.TriggerType != "USERINPUT" {
			continue
		}

		extra += 1
	}

	if ValidateFinished(ctx, extra, *workflowExecution) {
		return []string{}
	}

	for _, branch := range workflowExecution.Workflow.Branches {
		// Check what the parent is first. If it's trigger - skip
		sourceFound := false
		destinationFound := false
		for _, action := range workflowExecution.Workflow.Actions {
			if action.ID == branch.SourceID {
				sourceFound = true
			}

			if action.ID == branch.DestinationID {
				destinationFound = true
			}
		}

		if !sourceFound || !destinationFound {
			for _, trigger := range workflowExecution.Workflow.Triggers {
				if trigger.ID == branch.SourceID {
					sourceFound = true
				}

				if trigger.ID == branch.DestinationID {
					destinationFound = true
				}
			}
		}

		foundCnt := 0
		if sourceFound {
			parents[branch.DestinationID] = append(parents[branch.DestinationID], branch.SourceID)
			foundCnt += 1
		}

		if destinationFound {
			children[branch.SourceID] = append(children[branch.SourceID], branch.DestinationID)
			foundCnt += 1
		}

		if foundCnt != 2 {
			log.Printf("[INFO] Missing branch fullfillment!")
		}
	}

	nextActions = findMissingChildren(ctx, workflowExecution, children, inputNode, []string{})

	//log.Printf("[DEBUG][%s] Checking what are next actions in workflow %s. Results: %d/%d. NextActions: %s", workflowExecution.ExecutionId, workflowExecution.ExecutionId, len(workflowExecution.Results), len(workflowExecution.Workflow.Actions)+extra, nextActions)

	return nextActions
}

// Decideds what should happen next. Used both for cloud & onprem environments
// Added early 2023 as yet another way to standardize decisionmaking of app executions
func DecideExecution(ctx context.Context, workflowExecution WorkflowExecution, environment string) (WorkflowExecution, []Action) {
	// ensuring always latest
	newexec, err := GetWorkflowExecution(ctx, workflowExecution.ExecutionId)
	if err != nil {
		log.Printf("[ERROR] Failed to get workflow execution in Decide: %s", err)
	} else {
		workflowExecution = *newexec
	}

	startAction, extra, children, parents, visited, executed, nextActions, environments := GetExecutionVariables(ctx, workflowExecution.ExecutionId)
	if len(startAction) == 0 {
		startAction = workflowExecution.Start
		if len(startAction) == 0 {
			log.Printf("[WARNING] Didn't find execution start action. Setting it to workflow start action.")
			startAction = workflowExecution.Workflow.Start
		}
	}

	// Dedup results just in case
	newResults := []ActionResult{}
	handled := []string{}
	for _, result := range workflowExecution.Results {
		if ArrayContains(handled, result.Action.ID) {
			continue
		}

		handled = append(handled, result.Action.ID)
		newResults = append(newResults, result)
	}

	workflowExecution.Results = newResults
	relevantActions := []Action{}

	log.Printf("[INFO][%s] Inside Decide execution with %d / %d results (extra: %d)", workflowExecution.ExecutionId, len(workflowExecution.Results), len(workflowExecution.Workflow.Actions)+extra, extra)

	if len(startAction) == 0 {
		startAction = workflowExecution.Start

		if len(startAction) == 0 {
			log.Printf("[WARNING] Didn't find execution start action. Setting it to workflow start action (%s)", workflowExecution.Workflow.Start)
			startAction = workflowExecution.Workflow.Start
			workflowExecution.Start = workflowExecution.Workflow.Start
		}
	}

	//log.Printf("[DEBUG] NEXTACTIONS: %s", nextActions)
	//if len(nextActions) == 0 {
	//	nextActions = append(nextActions, startAction)
	//}

	queueNodes := []string{}
	if len(workflowExecution.Results) == 0 {
		nextActions = []string{startAction}
	} else {
		// This is to re-check the nodes that exist and whether they should continue
		appendActions := []string{}
		for _, item := range workflowExecution.Results {

			// FIXME: Check whether the item should be visited or not
			// Do the same check as in walkoff.go - are the parents done?
			// If skipped and both parents are skipped: keep as skipped, otherwise queue
			if item.Status == "SKIPPED" {
				isSkipped := true

				for _, branch := range workflowExecution.Workflow.Branches {
					// 1. Finds branches where the destination is our node
					// 2. Finds results of those branches, and sees the status
					// 3. If the status isn't skipped or failure, then it will still run this node
					if branch.DestinationID == item.Action.ID {
						for _, subresult := range workflowExecution.Results {
							if subresult.Action.ID == branch.SourceID {
								if subresult.Status != "SKIPPED" && subresult.Status != "FAILURE" {
									//log.Printf("\n\n\nSUBRESULT PARENT STATUS: %s\n\n\n", subresult.Status)
									isSkipped = false

									break
								}
							}
						}
					}
				}

				if isSkipped {
					//log.Printf("Skipping %s as all parents are done", item.Action.Label)
					if !ArrayContains(visited, item.Action.ID) {
						//log.Printf("[INFO] Adding visited (1): %s", item.Action.Label)
						visited = append(visited, item.Action.ID)
					}
				} else {
					//log.Printf("[INFO] Continuing %s as all parents are NOT done", item.Action.Label)
					// FIXME: Remove this  visited?
					//visited = append(visited, item.Action.ID)

					appendActions = append(appendActions, item.Action.ID)
				}
			} else {
				if item.Status == "FINISHED" {
					//log.Printf("[INFO] Adding visited (2): %s", item.Action.Label)
					visited = append(visited, item.Action.ID)
				}
			}

			//if len(nextActions) == 0 {
			//nextActions = append(nextActions, children[item.Action.ID]...)
			for _, child := range children[item.Action.ID] {
				if !ArrayContains(nextActions, child) && !ArrayContains(visited, child) && !ArrayContains(visited, child) {
					nextActions = append(nextActions, child)
				}
			}

			if len(appendActions) > 0 {
				//log.Printf("APPENDED NODES: %s", appendActions)
				nextActions = append(nextActions, appendActions...)
			}
		}
	}

	//log.Printf("Nextactions: %s", nextActions)
	// This is a backup in case something goes wrong in this complex hellhole.
	// Max default execution time is 5 minutes for now anyway, which should take
	// care if it gets stuck in a loop.
	// FIXME: Force killing a worker should result in a notification somewhere
	if len(nextActions) == 0 {
		log.Printf("[DEBUG] No next action. Finished? Result vs Actions: %d - %d", len(workflowExecution.Results), len(workflowExecution.Workflow.Actions))
		extra = 0

		for _, trigger := range workflowExecution.Workflow.Triggers {
			if (trigger.Name == "User Input" && trigger.AppName == "User Input") || (trigger.Name == "Shuffle Workflow" && trigger.AppName == "Shuffle Workflow") {
				extra += 1
			}
		}

		exit := true
		for _, item := range workflowExecution.Results {
			if item.Status == "EXECUTING" {
				exit = false
				break
			}
		}

		if len(environments) == 1 {
			log.Printf("[INFO] Should send results to the backend because environments are %s", environments)
			ValidateFinished(ctx, extra, workflowExecution)
		}

		if exit && len(workflowExecution.Results) == len(workflowExecution.Workflow.Actions) {
			ValidateFinished(ctx, extra, workflowExecution)
			//handleAbortExecution(ctx, workflowExecution)
			return workflowExecution, relevantActions
		}

		// Look for the NEXT missing action
		notFound := []string{}
		for _, action := range workflowExecution.Workflow.Actions {
			found := false
			for _, result := range workflowExecution.Results {
				if action.ID == result.Action.ID {
					found = true
					break
				}
			}

			if !found {
				notFound = append(notFound, action.ID)
			}
		}

		//log.Printf("SOMETHING IS MISSING!: %s", notFound)
		for _, item := range notFound {
			if ArrayContains(executed, item) {
				log.Printf("%s has already executed but no result!", item)
				//return workflowExecution
			}

			// Visited means it's been touched in any way.
			outerIndex := -1
			for index, visit := range visited {
				if visit == item {
					outerIndex = index
					break
				}
			}

			if outerIndex >= 0 {
				log.Printf("Removing index %s from visited", item)
				visited = append(visited[:outerIndex], visited[outerIndex+1:]...)
			}

			fixed := 0
			for _, parent := range parents[item] {
				parentResult := ActionResult{}
				workflowExecution, parentResult = GetResult(ctx, workflowExecution, parent)
				if parentResult.Status == "FINISHED" || parentResult.Status == "SUCCESS" || parentResult.Status == "SKIPPED" || parentResult.Status == "FAILURE" {
					fixed += 1
				}
			}

			if fixed == len(parents[item]) {
				nextActions = append(nextActions, item)
			}

			// If it's not executed and not in nextActions
		}
	}

	//log.Printf("Checking nextactions: %s", nextActions)
	for _, node := range nextActions {
		nodeChildren := children[node]
		for _, child := range nodeChildren {
			if !ArrayContains(queueNodes, child) {
				queueNodes = append(queueNodes, child)
			}
		}
	}

	// IF NOT VISITED && IN toExecuteOnPrem
	// SKIP if it's not onprem
	for _, nextAction := range nextActions {
		//log.Printf("[DEBUG] Handling nextAction %s", nextAction)
		action := GetAction(workflowExecution, nextAction, environment)

		// Using cache to ensure the same app isn't ran twice
		// May arise due to one app being nanoseconds before another

		// Not really sure how this edgecase happens.

		// FIXME
		// Execute, as we don't really care if env is not set? IDK
		if action.Environment != environment { //&& action.Environment != "" {
			if strings.ToLower(action.Environment) == strings.ToLower(environment) {
				// Fixing names
				action.Environment = environment
			} else {
				//log.Printf("envs: %s", environments)
				//log.Printf("[WARNING] Bad environment for node: %s. Want %s", action.Environment, environment)
				action.Environment = "cloud"
				//DeleteCache(ctx, newExecId)
				//continue
			}
		}

		// check whether the parent is finished executing

		fixed := 0
		continueOuter := true
		if action.IsStartNode {
			continueOuter = false
		} else if len(parents[nextAction]) > 0 {
			// Wait for parents to finish executing
			childNodes := FindChildNodes(workflowExecution, nextAction, []string{}, []string{})
			for _, parent := range parents[nextAction] {
				// Check if the parent is also a child. This can ensure continueation no matter what
				if ArrayContains(childNodes, parent) {
					log.Printf("[ERROR][%s] Parent %s is also a child of %s. Skipping parent check", workflowExecution.ExecutionId, parent, nextAction)
					fixed += 1
					continue
				}

				_, parentResult := GetResult(ctx, workflowExecution, parent)
				if parentResult.Status == "FINISHED" || parentResult.Status == "SUCCESS" || parentResult.Status == "SKIPPED" || parentResult.Status == "FAILURE" {
					fixed += 1
				} else {
					log.Printf("[WARNING][%s] Parentstatus for node %s is '%s' - NOT success, finished or skipped", workflowExecution.ExecutionId, parent, parentResult.Status)
					// Should check if it's actually RAN at all?

					parentId := fmt.Sprintf("%s_%s", workflowExecution.ExecutionId, parent)
					_, err := GetCache(ctx, parentId)
					if err != nil {
						//log.Printf("[INFO] No cache for parent ID %#v", parentId)
					} else {
						//log.Printf("Parent ID already ran. How long ago?")
					}
				}
			}

			if fixed == len(parents[nextAction]) {
				continueOuter = false
			}
		} else {
			log.Printf("[INFO] No parents for %s", action.Label)
			continueOuter = false
		}

		if continueOuter {
			// FIXME: Get the result here from cache in case it exists?
			log.Printf("[DEBUG][%s] Parents of %s (%s) aren't finished yet (%d/%d). Parents: %s", workflowExecution.ExecutionId, action.Label, nextAction, fixed, len(parents[nextAction]), strings.Join(parents[nextAction], ", "))
			//for _, parent := range parents[nextAction] {
			//	_, parentResult := GetResult(ctx, workflowExecution, parent)
			//}
			// I think this is where we rerun.

			continue
		} else {
			//log.Printf("\n\n[DEBUG][%s] ALL Parents of %s (%s) are finished. (%d/%d): %s. But are they succeeded?", workflowExecution.ExecutionId, action.Label, nextAction, fixed, len(parents[nextAction]), strings.Join(parents[nextAction], ", "))

		}

		//if action.Label == "Add_to_cache" {
		//	log.Printf("\n\n\nFOUND ACTION!!\n\n\n")
		//}

		// get action status
		workflowExecution, actionResult := GetResult(ctx, workflowExecution, nextAction)
		if actionResult.Action.ID == action.ID {
			//log.Printf("\n\n[INFO] %s (%s) already has status %s\n\n", action.Label, action.ID, actionResult.Status)
			//DeleteCache(ctx, newExecId)
			continue
		} else {
			//log.Printf("[INFO] %s:%s has no status result yet. Should execute.", action.Name, action.ID)
		}

		newExecId := fmt.Sprintf("%s_%s", workflowExecution.ExecutionId, nextAction)
		_, err := GetCache(ctx, newExecId)
		if err == nil {
			//log.Printf("\n\n[DEBUG] Already found %s - returning\n\n", newExecId)
			continue
		}

		//cacheData := []byte("1")
		//err = SetCache(ctx, newExecId, cacheData, 2)
		//if err != nil {
		//	log.Printf("[WARNING] Failed setting cache for action %s: %s", newExecId, err)
		//} else {
		//}

		if action.AppName == "Shuffle Workflow" {
			branchesFound := 0
			parentFinished := 0

			for _, item := range workflowExecution.Workflow.Branches {
				if item.DestinationID == action.ID {
					branchesFound += 1

					for _, result := range workflowExecution.Results {
						if result.Action.ID == item.SourceID {
							// Check for fails etc
							if result.Status == "SUCCESS" || result.Status == "SKIPPED" {
								parentFinished += 1
							} else {
								log.Printf("Parent %s has status %s", result.Action.Label, result.Status)
							}

							break
						}
					}
				}
			}

			//log.Printf("[DEBUG] Should execute %s (?). Branches: %d. Parents done: %d", action.AppName, branchesFound, parentFinished)
			if branchesFound == parentFinished {
				action.Environment = environment
				action.AppName = "shuffle-subflow"
				action.Name = "run_subflow"
				action.AppVersion = "1.0.0"

				//appname := action.AppName
				//appversion := action.AppVersion
				//appname = strings.Replace(appname, ".", "-", -1)
				//appversion = strings.Replace(appversion, ".", "-", -1)

				//visited = append(visited, action.ID)
				//executed = append(executed, action.ID)

				trigger := Trigger{}
				for _, innertrigger := range workflowExecution.Workflow.Triggers {
					if innertrigger.ID == action.ID {
						trigger = innertrigger
						break
					}
				}

				// FIXME: Add startnode from frontend
				action.ExecutionDelay = trigger.ExecutionDelay
				action.Parameters = []WorkflowAppActionParameter{}
				for _, parameter := range trigger.Parameters {
					parameter.Variant = "STATIC_VALUE"
					action.Parameters = append(action.Parameters, parameter)
				}

				action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
					Name:  "source_workflow",
					Value: workflowExecution.Workflow.ID,
				})

				action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
					Name:  "source_execution",
					Value: workflowExecution.ExecutionId,
				})

				action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
					Name:  "source_auth",
					Value: workflowExecution.Authorization,
				})

				action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
					Name:  "source_node",
					Value: action.ID,
				})

				backendUrl := os.Getenv("BASE_URL")
				if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
					backendUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
				}

				if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
					backendUrl = os.Getenv("SHUFFLE_CLOUDRUN_URL")
				}

				if len(backendUrl) > 0 {
					action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
						Name:  "backend_url",
						Value: backendUrl,
					})
				}

			}
		} else if action.AppName == "User Input" {
			branchesFound := 0
			parentFinished := 0

			for _, item := range workflowExecution.Workflow.Branches {
				if item.DestinationID == action.ID {
					branchesFound += 1

					for _, result := range workflowExecution.Results {
						if result.Action.ID == item.SourceID {
							// Check for fails etc
							if result.Status == "SUCCESS" || result.Status == "SKIPPED" {
								parentFinished += 1
							} else {
								log.Printf("Parent %s has status %s", result.Action.Label, result.Status)
							}

							break
						}
					}
				}
			}

			//log.Printf("[DEBUG] Should execute parent %s (?). Branches: %d. Parents done: %d", action.AppName, branchesFound, parentFinished)
			if branchesFound == parentFinished {

				if action.ID == workflowExecution.Start {
					log.Printf("Skipping because it's the startnode")
					visited = append(visited, action.ID)
					executed = append(executed, action.ID)
					continue
				} else {
					//log.Printf("Should stop after this iteration because it's user-input based. %s", action)
					log.Printf("[DEBUG] Should stop after this iteration because it's user-input based.") //%s", action)

					trigger := Trigger{}
					for _, innertrigger := range workflowExecution.Workflow.Triggers {
						if innertrigger.ID == action.ID {
							trigger = innertrigger
							break
						}
					}

					trigger.LargeImage = ""
					triggerData, err := json.Marshal(trigger)
					if err != nil {
						log.Printf("[WARNING] Failed unmarshalling action: %s", err)
						triggerData = []byte("Failed unmarshalling. Cancel execution!")
					}

					_ = triggerData

					timeNow := int64(time.Now().Unix())
					result := ActionResult{
						Action:        action,
						ExecutionId:   workflowExecution.ExecutionId,
						Authorization: workflowExecution.Authorization,
						Result:        "{\"success\": true, \"reason\": \"WAITING FOR USER INPUT\"}",
						StartedAt:     timeNow,
						CompletedAt:   0,
						Status:        "WAITING",
					}

					workflowExecution.Results = append(workflowExecution.Results, result)
					workflowExecution.Status = "WAITING"
					err = SetWorkflowExecution(ctx, workflowExecution, true)
					if err != nil {
						log.Printf("[ERROR] Error saving workflow execution actionresult setting: %s", err)
						break
					}

					action.Environment = environment
					action.AppName = "shuffle-subflow"
					action.Name = "run_userinput"
					action.AppVersion = "1.1.0"

					for _, innertrigger := range workflowExecution.Workflow.Triggers {
						if innertrigger.ID == action.ID {
							trigger = innertrigger
							break
						}
					}

					smsEnabled := false
					emailEnabled := false
					subflowEnabled := false
					for _, trigger := range trigger.Parameters {
						if trigger.Name == "type" {
							if strings.Contains(trigger.Value, "sms") {
								smsEnabled = true
							}

							if strings.Contains(trigger.Value, "email") {
								emailEnabled = true
							}

							if strings.Contains(trigger.Value, "subflow") {
								subflowEnabled = true
							}
						}
					}

					//log.Printf("\n\nSHOULD RUN USER INPUT! SMS: %#v, email: %#v, subflow: %#v \n\n", smsEnabled, emailEnabled, subflowEnabled)

					// FIXME: Add startnode from frontend
					action.ExecutionDelay = trigger.ExecutionDelay
					action.Parameters = []WorkflowAppActionParameter{}
					for _, parameter := range trigger.Parameters {
						parameter.Variant = "STATIC_VALUE"
						if parameter.Name == "alertinfo" {
							parameter.Name = "information"
						}

						if parameter.Name == "sms" && smsEnabled == false {
							continue
						}

						if parameter.Name == "email" && emailEnabled == false {
							continue
						}

						if parameter.Name == "subflow" && subflowEnabled == false {
							continue
						}

						action.Parameters = append(action.Parameters, parameter)
					}

					//action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
					//	Name:  "source_workflow",
					//	Value: workflowExecution.Workflow.ID,
					//})

					//action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
					//	Name:  "source_execution",
					//	Value: workflowExecution.ExecutionId,
					//})

					//action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
					//	Name:  "source_auth",
					//	Value: workflowExecution.Authorization,
					//})

					action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
						Name:  "startnode",
						Value: workflowExecution.Start,
					})

					backendUrl := os.Getenv("BASE_URL")
					if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
						backendUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
					}

					if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
						backendUrl = os.Getenv("SHUFFLE_CLOUDRUN_URL")
					}

					// Fallback
					if len(backendUrl) == 0 {
						backendUrl = "https://shuffler.io"
					}

					if len(backendUrl) > 0 {
						action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
							Name:  "backend_url",
							Value: backendUrl,
						})
					}

					log.Printf("Starting with sourcenode '%s'", trigger.ID)
					action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
						Name:  "source_node",
						Value: trigger.ID,
					})

					// If sms/email, it should be setting the apikey based on the org
					syncApikey := workflowExecution.Authorization
					if project.Environment != "cloud" {
						org, err := GetOrg(ctx, workflowExecution.ExecutionOrg)
						if err == nil {
							log.Printf("Got syncconfig key: %s", org.SyncConfig.Apikey)
							syncApikey = org.SyncConfig.Apikey
						} else {
							log.Printf("[ERROR] Failed to get org %s: %s", workflowExecution.ExecutionOrg, err)
						}
					}

					action.Parameters = append(action.Parameters, WorkflowAppActionParameter{
						Name:  "user_apikey",
						Value: syncApikey,
					})

					/*
						for _, trigger := range trigger.Parameters {
							if trigger.Name == "email" {
								itemsplit := strings.Split(trigger.Value, ",")
								for _, mail := range itemsplit {
									users = append(users, mail)
								}
							} else if trigger.Name == "alertinfo" {
								mailbody = trigger.Value
							} else if trigger.Name == "type" {
								if strings.Contains(trigger.Value, "sms") {
									smsEnabled = true
								}

								if strings.Contains(trigger.Value, "email") {
									emailEnabled = true
								}
							} else if trigger.Name == "sms" {
								smsUser = trigger.Value
							} else if trigger.Name == "subflow" {
								log.Printf("[DEBUG] Should handle subflow in user input")
							}
						}
					*/

					// Still having personal to see if it works or not
					/*
						users := []string{}
						smsUser := ""
						mailbody := ""

						// FIXME: Change the startnode to NEXT one :thinking:

						startnode := workflowExecution.Start
						if emailEnabled {
							mailBody := Mailcheck{
								Targets:            users,
								Type:               "User input",
								WorkflowId:         workflowExecution.Workflow.ID,
								Start:              startnode,
								Body:               mailbody,
								ReferenceExecution: workflowExecution.ExecutionId,
								Subject:            "An alert from Shuffle",
							}
							_ = mailBody

							//err = sendAlertMail(ctx, mailBody, workflowExecution.ExecutionOrg)
							//if err != nil {
							//	log.Printf("[INFO] Failed sending email about user input: %s", err)
							//}
						}

						_ = smsUser
						_ = startAction
						if smsEnabled {
							if len(mailbody) > 140 {
								mailbody = mailbody[0:139]
							}

							url := "https://shuffler.io"

							continueUrl := fmt.Sprintf("%s/api/v1/workflows/%s/execute?authorization=%s&start=%s&reference_execution=%s&answer=true", url, workflowExecution.Workflow.ID, workflowExecution.Authorization, startnode, workflowExecution.ExecutionId)
							stopUrl := fmt.Sprintf("%s/api/v1/workflows/%s/execute?authorization=%s&start=%s&reference_execution=%s&answer=false", url, workflowExecution.Workflow.ID, workflowExecution.Authorization, startnode, workflowExecution.ExecutionId)

							mailbody += fmt.Sprintf("\n\nContinue: %s\n\nStop: %s", continueUrl, stopUrl)

							//err = sendSMS(ctx, mailbody, []string{smsUser}, workflowExecution.Workflow.ExecutingOrg.Id)
							//if err != nil {
							//	log.Printf("[INFO] Failed sending sms about user input: %s", err)
							//}
						}

						break
					*/
				}
			}
		} else {
			//log.Printf("Handling action %s", action)
		}

		// Here it's still in a loop..?
		_, _, _, _, _, executed, _, _ = GetExecutionVariables(ctx, workflowExecution.ExecutionId)
		if ArrayContains(visited, action.ID) || ArrayContains(executed, action.ID) {
			log.Printf("[WARNING][%s] SKIP EXECUTION %s:%s with label %s", workflowExecution.ExecutionId, action.AppName, action.AppVersion, action.Label)
			continue
		} else {
			// FIXME? This was a test to check if a result was finished or not after a certain time. Not viable for production (obv)

			//time.Sleep(1 * time.Second)
			//validateExecution, err := shuffle.GetWorkflowExecution(ctx, workflowExecution.ExecutionId)
			//log.Printf("Got execution with %d results", len(validateExecution.Results))
			//if err == nil {
			//	skipAction := false
			//	for _, result := range validateExecution.Results {
			//		if result.Action.ID == action.ID {
			//			skipAction = true
			//			break
			//		}
			//	}

			//	if skipAction {
			//		log.Printf("[DEBUG] Skipping action %s afterall.", action.Label)
			//		continue
			//	}
			//}
		}

		// Verify if parents are done

		log.Printf("[INFO][%s] Should execute %s:%s (%s) with label %s", workflowExecution.ExecutionId, action.AppName, action.AppVersion, action.ID, action.Label)
		relevantActions = append(relevantActions, action)
	}

	return workflowExecution, relevantActions
}

func GetExternalClient(baseUrl string) *http.Client {
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")

	transport := http.DefaultTransport.(*http.Transport)
	transport.MaxIdleConnsPerHost = 100
	transport.ResponseHeaderTimeout = time.Second * 60
	transport.Proxy = nil

	skipSSLVerify := false
	if strings.ToLower(os.Getenv("SHUFFLE_OPENSEARCH_SKIPSSL_VERIFY")) == "true" {
		log.Printf("[DEBUG] SKIPPING SSL verification with Opensearch")
		skipSSLVerify = true
	}

	transport.TLSClientConfig = &tls.Config{
		MinVersion:         tls.VersionTLS11,
		InsecureSkipVerify: skipSSLVerify,
	}

	//getStats()

	if (len(httpProxy) > 0 || len(httpsProxy) > 0) && baseUrl != "http://shuffle-backend:5001" {
		//client = &http.Client{}
	} else {
		if len(httpProxy) > 0 {
			log.Printf("[INFO] Running with HTTP proxy %s (env: HTTP_PROXY)", httpProxy)

			url_i := url.URL{}
			url_proxy, err := url_i.Parse(httpProxy)
			if err == nil {
				transport.Proxy = http.ProxyURL(url_proxy)
			}
		}
		if len(httpsProxy) > 0 {
			log.Printf("[INFO] Running with HTTPS proxy %s (env: HTTPS_PROXY)", httpsProxy)

			url_i := url.URL{}
			url_proxy, err := url_i.Parse(httpsProxy)
			if err == nil {
				transport.Proxy = http.ProxyURL(url_proxy)
			}
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 60,
	}

	return client
}

func GetAllAppCategories() []AppCategory {
	categories := []AppCategory{
		AppCategory{
			Name:         "Cases",
			Color:        "",
			Icon:         "cases",
			ActionLabels: []string{"Create ticket", "List tickets", "Get ticket", "Create ticket", "Close ticket", "Add comment", "Update ticket"},
			RequiredFields: map[string][]string{
				"Create ticket": []string{"title"},
				"Add comment":   []string{"comment"},
				"Lis tickets":   []string{"time_range"},
			},
			OptionalFields: map[string][]string{
				"Create ticket": []string{"description"},
			},
		},
		AppCategory{
			Name:           "Communication",
			Color:          "",
			Icon:           "communication",
			ActionLabels:   []string{"List Messages", "Send Message", "Get Message", "Search messages"},
			RequiredFields: map[string][]string{},
			OptionalFields: map[string][]string{},
		},
		AppCategory{
			Name:           "SIEM",
			Color:          "",
			Icon:           "siem",
			ActionLabels:   []string{"Search", "List Alerts", "Close Alert", "Create detection", "Add to lookup list"},
			RequiredFields: map[string][]string{},
			OptionalFields: map[string][]string{},
		},
		AppCategory{
			Name:           "Eradication",
			Color:          "",
			Icon:           "eradication",
			ActionLabels:   []string{"List Alerts", "Close Alert", "Create detection", "Block hash", "Search Hosts", "Isolate host", "Unisolate host"},
			RequiredFields: map[string][]string{},
			OptionalFields: map[string][]string{},
		},
		AppCategory{
			Name:           "Assets",
			Color:          "",
			Icon:           "assets",
			ActionLabels:   []string{"List Assets", "Get Asset", "Search Assets", "Search Users", "Search endpoints", "Search vulnerabilities"},
			RequiredFields: map[string][]string{},
			OptionalFields: map[string][]string{},
		},
		AppCategory{
			Name:           "Intel",
			Color:          "",
			Icon:           "intel",
			ActionLabels:   []string{"Get IOC", "Search IOC", "Create IOC", "Update IOC", "Delete IOC"},
			RequiredFields: map[string][]string{},
			OptionalFields: map[string][]string{},
		},
		AppCategory{
			Name:           "IAM",
			Color:          "",
			Icon:           "iam",
			ActionLabels:   []string{"Reset Password", "Enable user", "Disable user", "Get Identity", "Get Asset", "Search Identity"},
			RequiredFields: map[string][]string{},
			OptionalFields: map[string][]string{},
		},
		AppCategory{
			Name:           "Network",
			Color:          "",
			Icon:           "network",
			ActionLabels:   []string{"Get Rules", "Allow IP", "Block IP"},
			RequiredFields: map[string][]string{},
			OptionalFields: map[string][]string{},
		},
		AppCategory{
			Name:           "Other",
			Color:          "",
			Icon:           "other",
			ActionLabels:   []string{"Update Info", "Get Info", "Get Status", "Get Version", "Get Health", "Get Config", "Get Configs", "Get Configs by type", "Get Configs by name", "Run script"},
			RequiredFields: map[string][]string{},
			OptionalFields: map[string][]string{},
		},
	}

	return categories
}

// Function with the name RemoveFromArray to remove a string from a string array
func RemoveFromArray(array []string, element string) []string {
	for i, v := range array {
		if v == element {
			return append(array[:i], array[i+1:]...)
		}
	}

	return array
}

func FindRelevantApps(appname string, apps []WorkflowApp) []WorkflowApp {
	return []WorkflowApp{}
}

func FindMatchingCategoryApps(category string, apps []WorkflowApp, org *Org) []WorkflowApp {
	if len(category) == 0 {
		return []WorkflowApp{}
	}

	category = strings.ToLower(category)
	parsedCategories := map[string]Category{
		"siem":          org.SecurityFramework.SIEM,
		"email":         org.SecurityFramework.Communication,
		"communication": org.SecurityFramework.Communication,
		"assets":        org.SecurityFramework.Assets,
		"cases":         org.SecurityFramework.Cases,
		"network":       org.SecurityFramework.Network,
		"intel":         org.SecurityFramework.Intel,
		"eradication":   org.SecurityFramework.EDR,
		"edr":           org.SecurityFramework.EDR,
		"iam":           org.SecurityFramework.IAM,
	}

	var matchingApps []WorkflowApp
	foundCategory, ok := parsedCategories[category]
	if ok && len(foundCategory.Name) > 0 {
		parsedCategoryNames := strings.Split(foundCategory.Name, ",")
		for _, catApp := range parsedCategoryNames {
			catApp = strings.ToLower(strings.TrimSpace(catApp))

			found := false
			for _, app := range apps {
				if strings.ToLower(app.Name) == catApp {
					matchingApps = append(matchingApps, app)
					found = true
					break
				}
			}

			if !found {
				log.Printf("[INFO] Could not find app %s in category %s", catApp, category)
			}
		}
	}

	log.Printf("[INFO] Found %d apps in category '%s'", len(matchingApps), category)
	if category == "email" {
		category = "communication"
	}

	for _, app := range apps {
		if len(app.Categories) == 0 {
			continue
		}

		if strings.ToLower(app.Categories[0]) != category {
			continue
		}

		matchingApps = append(matchingApps, app)
	}

	return matchingApps
}

// Apps to test:
// Communication: Slack & Teams -> Send message
// Ticketing: TheHive & Jira 		-> Create ticket

// 1. Check the category, if it has the label in it
// 2. Get the app (openapi & app config)
// 3. Check if the label exists for the app in any action
// 4. Do the translation~ (workflow or direct)
// 5. Run app action/workflow
// 6. Return result from the app
func RunCategoryAction(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Just here to verify that the user is logged in
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[AUDIT] Api authentication failed in run category action: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to run category action: %s (%s)", user.Username, user.Id)
		resp.WriteHeader(403)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[WARNING] Error with body read for category action: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var value CategoryAction
	err = json.Unmarshal(body, &value)
	if err != nil {
		log.Printf("[WARNING] Error with unmarshal tmpBody in category action: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)

	categories := GetAllAppCategories()

	value.Category = strings.ToLower(value.Category)
	value.Label = strings.ReplaceAll(strings.ToLower(value.Label), " ", "_")
	value.AppName = strings.ReplaceAll(strings.ToLower(value.AppName), " ", "_")

	foundIndex := -1
	labelIndex := -1
	if len(value.Category) > 0 {

		for categoryIndex, category := range categories {
			if strings.ToLower(category.Name) != value.Category {
				continue
			}

			foundIndex = categoryIndex
			break
		}

		if foundIndex >= 0 {
			for _, label := range categories[foundIndex].ActionLabels {
				if strings.ReplaceAll(strings.ToLower(label), " ", "_") == value.Label {
					labelIndex = foundIndex
				}
			}
		}
	} else {
		log.Printf("[DEBUG] No category set in category action")
	}

	log.Printf("[INFO] Found label %s in category %d with label %d", value.Label, foundIndex, labelIndex)

	newapps, err := GetPrioritizedApps(ctx, user)
	if err != nil {
		log.Printf("[WARNING] Failed getting apps in category action: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Failed loading apps. Contact support@shuffler.io"}`))
		return
	}

	selectedApp := WorkflowApp{}
	selectedCategory := AppCategory{}
	selectedAction := WorkflowAppAction{}
	for _, app := range newapps {
		if app.Name == "" || len(app.Categories) == 0 {
			continue
		}

		if app.ID == value.AppName || strings.ReplaceAll(strings.ToLower(app.Name), " ", "_") == value.AppName {
			log.Printf("[DEBUG] Found app: %s vs %s (%s)", app.Name, value.AppName, app.ID)
			selectedApp = app

			for _, action := range app.Actions {
				if len(action.CategoryLabel) == 0 {
					continue
				} else {
					log.Printf("[DEBUG] Found category label %s", action.CategoryLabel)
				}

				if len(value.ActionName) > 0 && action.Name != value.ActionName {
					continue
				}

				// For now just finding the first one
				//log.Printf("Hi: %s & %s", strings.ReplaceAll(strings.ToLower(action.CategoryLabel[0]), " ", "_"), strings.ReplaceAll(strings.ToLower(value.Label), " ", "_"))

				if strings.ReplaceAll(strings.ToLower(action.CategoryLabel[0]), " ", "_") == strings.ReplaceAll(strings.ToLower(value.Label), " ", "_") {

					selectedAction = action

					for _, category := range categories {
						if strings.ToLower(category.Name) == strings.ToLower(app.Categories[0]) {
							selectedCategory = category
							break
						}
					}

					log.Printf("[INFO] Found label %s in app %s. ActionName: %s", value.Label, app.Name, action.Name)
					break
				}
			}

			break
		}
	}

	if len(selectedApp.ID) == 0 {
		log.Printf("[WARNING] Couldn't find app with ID or name '%s' active in org %s (%s)", value.AppName, user.ActiveOrg.Name, user.ActiveOrg.Id)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Failed finding an app with that name or ID"}`))
		return
	}

	if len(selectedAction.Name) == 0 {
		log.Printf("[WARNING] Couldn't find the label '%s' in app '%s'", value.Label, selectedApp.Name)
		resp.WriteHeader(500)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed finding the label in app '%s'"}`, selectedApp.Name)))
		return
	}

	log.Printf("[INFO] Got action: %#v. Required bodyfields: %#v", selectedAction.Name, selectedAction.RequiredBodyFields)
	// Need translation here, now that we have the action
	// Should just do an app injection into the workflow?

	foundFields := 0
	missingFields := []string{}
	baseFields := []string{}
	_ = foundFields
	for innerLabel, labelValue := range selectedCategory.RequiredFields {
		if strings.ReplaceAll(strings.ToLower(innerLabel), " ", "_") != value.Label {
			continue
		}

		baseFields = labelValue
		missingFields = labelValue
		for _, field := range value.Fields {
			if ArrayContains(labelValue, field.Key) {
				// Remove from missingFields
				missingFields = RemoveFromArray(missingFields, field.Key)

				foundFields += 1
			}
		}

		break
	}

	// FIXME: Check if ALL fields for the target app can be fullfiled
	// E.g. for Jira: Org_id is required.

	if foundFields != len(baseFields) {
		log.Printf("[WARNING] Not all required fields were found in category action. Want: %#v, have: %#v", baseFields, value.Fields)
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Not all required fields are set", "label": "%s", "missing_fields": "%s"}`, value.Label, strings.Join(missingFields, ","))))
		return

	}

	/*











	 */

	// 1. Generate the workflow based on the input above (make sure to set it to hidden)
	// 2. Run the workflow
	// 3. Get results for it
	// 4. Delete it

	standardType := "cases"

	startId := uuid.NewV4().String()
	newWorkflow := Workflow{
		ID:        uuid.NewV4().String(),
		Name:      fmt.Sprintf("Category action: %s", selectedAction.Name),
		Generated: true,
		Hidden:    true,
		Start:     startId,
		OrgId:     user.ActiveOrg.Id,
	}

	environment := "Cloud"
	// FIXME: Make it dynamic based on the default env
	if project.Environment != "cloud" {
		environment = "Shuffle"
	}

	startAction := Action{
		Name:        "get_standardized_data",
		Label:       "Get standardized data",
		AppName:     "Shuffle Tools",
		AppVersion:  "1.2.0",
		AppID:       "3e2bdf9d5069fe3f4746c29d68785a6a",
		Environment: environment,
		IsStartNode: true,
		ID:          startId,
		Parameters: []WorkflowAppActionParameter{
			WorkflowAppActionParameter{
				Name:      "json_input",
				Value:     "$exec",
				Multiline: true,
				Variant:   "STATIC_VALUE",
			},
			WorkflowAppActionParameter{
				Name:  "input_type",
				Value: standardType,
				Options: []string{
					"cases",
				},
				Variant: "STATIC_VALUE",
			},
		},
		Position: Position{
			X: 0.0,
			Y: 0.0,
		},
	}

	log.Printf("\n\n[INFO] Added workflow %s\n\n", newWorkflow.ID)
	newWorkflow.Actions = append(newWorkflow.Actions, startAction)

	foundAuthenticationId := value.AuthenticationId

	// Sample. Point is to get authentication and choose the latest otherwise
	if foundAuthenticationId == "" {
		foundAuthenticationId = "378d651f-6379-4c25-aebf-2be55af06825"
	}

	refUrl := ""
	if project.Environment == "cloud" {
		location := "europe-west2"
		if len(os.Getenv("SHUFFLE_GCEPROJECT_REGION")) > 0 {
			location = os.Getenv("SHUFFLE_GCEPROJECT_REGION")
		}

		gceProject := os.Getenv("SHUFFLE_GCEPROJECT")
		functionName := fmt.Sprintf("%s-%s", selectedApp.Name, selectedApp.ID)
		functionName = strings.ToLower(strings.Replace(strings.Replace(strings.Replace(strings.Replace(strings.Replace(functionName, "_", "-", -1), ":", "-", -1), "-", "-", -1), " ", "-", -1), ".", "-", -1))

		refUrl = fmt.Sprintf("https://%s-%s.cloudfunctions.net/%s", location, gceProject, functionName)
	}

	// Now start connecting it to the correct app (Jira?)
	secondAction := Action{
		Name:        selectedAction.Name,
		Label:       selectedAction.Name,
		AppName:     selectedApp.Name,
		AppVersion:  selectedApp.AppVersion,
		AppID:       selectedApp.ID,
		Environment: environment,
		IsStartNode: false,
		ID:          uuid.NewV4().String(),
		Parameters:  []WorkflowAppActionParameter{},
		Position: Position{
			X: 100.0,
			Y: 100.0,
		},
		AuthenticationId: foundAuthenticationId,
		IsValid:          true,
		ReferenceUrl:     refUrl,
	}

	log.Printf("Required bodyfields: %#v", selectedAction.RequiredBodyFields)
	missingFields = []string{}
	for _, param := range selectedAction.Parameters {
		// Optional > Required
		for _, field := range value.OptionalFields {
			if strings.ReplaceAll(strings.ToLower(field.Key), " ", "_") == strings.ReplaceAll(strings.ToLower(param.Name), " ", "_") {
				param.Value = field.Value
			}
		}

		// Parse input params into the field
		for _, field := range value.Fields {
			//log.Printf("%s & %s", strings.ReplaceAll(strings.ToLower(field.Key), " ", "_"), strings.ReplaceAll(strings.ToLower(param.Name), " ", "_"))
			if strings.ReplaceAll(strings.ToLower(field.Key), " ", "_") == strings.ReplaceAll(strings.ToLower(param.Name), " ", "_") {
				param.Value = field.Value
			}
		}

		if param.Required && len(param.Value) == 0 && !param.Configuration {
			//log.Printf("[INFO] Required - Key: %s, Value: %#v", param.Name, param.Value)
			missingFields = append(missingFields, param.Name)
		}

		secondAction.Parameters = append(secondAction.Parameters, param)
	}

	if len(missingFields) > 0 {
		log.Printf("[WARNING] Not all required fields were found in category action. Want: %#v", missingFields)
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Not all required fields are set", "label": "%s", "missing_fields": "%s"}`, value.Label, strings.Join(missingFields, ","))))
		return
	}

	// Add params and such of course
	newWorkflow.Actions = append(newWorkflow.Actions, secondAction)

	newWorkflow.Branches = append(newWorkflow.Branches, Branch{
		ID:            uuid.NewV4().String(),
		SourceID:      startAction.ID,
		DestinationID: secondAction.ID,
		Label:         "",
		HasError:      false,
		Decorator:     false,
		Conditions:    []Condition{},
	})

	// Set workflow in cache only?
	// That way it can be loaded during execution

	workflowBytes, err := json.Marshal(newWorkflow)
	if err != nil {
		log.Printf("[WARNING] Failed marshal of workflow during cache setting: %s", err)
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed packing workflow. Please try again and contact support if this persists."}`)))
		return
	}

	err = SetCache(ctx, fmt.Sprintf("workflow_%s", newWorkflow.ID), workflowBytes, 10)
	if err != nil {
		log.Printf("[WARNING] Failed setting cache for workflow during category run: %s", err)

		SetWorkflow(ctx, newWorkflow, newWorkflow.ID)
	}

	log.Printf("[DEBUG] Done preparing workflow %s (%s) to be ran for category action %s", newWorkflow.Name, newWorkflow.ID, selectedAction.Name)

	/*














	 */

	execData := ExecutionStruct{
		Start:             startId,
		ExecutionSource:   "category_action",
		ExecutionArgument: `{"title": "hello"}`,
	}

	newExecBody, _ := json.Marshal(execData)

	// Starting execution
	streamUrl := fmt.Sprintf("https://shuffler.io")
	if len(os.Getenv("SHUFFLE_GCEPROJECT")) > 0 && len(os.Getenv("SHUFFLE_GCEPROJECT_LOCATION")) > 0 {
		streamUrl = fmt.Sprintf("https://%s.%s.r.appspot.com", os.Getenv("SHUFFLE_GCEPROJECT"), os.Getenv("SHUFFLE_GCEPROJECT_LOCATION"))
	}

	if len(os.Getenv("SHUFFLE_CLOUDRUN_URL")) > 0 {
		streamUrl = fmt.Sprintf("%s", os.Getenv("SHUFFLE_CLOUDRUN_URL"))
	}

	streamUrl = fmt.Sprintf("%s/api/v1/workflows/%s/execute", streamUrl, newWorkflow.ID)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(
		"POST",
		streamUrl,
		bytes.NewBuffer(newExecBody),
	)
	if err != nil {
		log.Printf("[WARNING] Error in new request for execute generated workflow: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed preparing new request. Contact support."}`)))
		return
	}

	req.Header.Add("Authorization", request.Header.Get("Authorization"))
	newresp, err := client.Do(req)
	if err != nil {
		log.Printf("[WARNING] Error running body for execute generated workflow: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed running generated app. Contact support."}`)))
		return
	}

	executionBody, err := ioutil.ReadAll(newresp.Body)
	if err != nil {
		log.Printf("[WARNING] Failed reading body for execute generated workflow: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unmarshalling app response. Contact support."}`)))
		return
	}

	workflowExecution := WorkflowExecution{}
	err = json.Unmarshal(executionBody, &workflowExecution)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshalling body for execute generated workflow: %s", err)
	}

	returnBody := HandleRetValidation(ctx, workflowExecution, len(newWorkflow.Actions))
	log.Printf("[INFO] Got response from workflow: %#v", string(returnBody.Result))

	resp.WriteHeader(200)
	resp.Write([]byte(returnBody.Result))

}

func GetActiveCategories(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	// Just here to verify that the user is logged in
	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[AUDIT] Api authentication failed in run category action: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed authentication"}`))
		return
	}

	ctx := GetContext(request)

	//log.Printf("[INFO] Found label %s in category %d with label %d", value.Label, foundIndex, labelIndex)

	newapps, err := GetPrioritizedApps(ctx, user)
	if err != nil {
		log.Printf("[WARNING] Failed getting apps in category action: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Failed loading apps. Contact support@shuffler.io"}`))
		return
	}

	categories := GetAllAppCategories()
	//AppCategory{
	//	Name:         "Cases",
	//	Color:        "",
	//	Icon:         "cases",
	//	ActionLabels: []string{"Create ticket"},
	//	LabelApps: []AppCategoryLabel{
	//		AppCategoryLabel{
	//			Name: "Create ticket",
	//			Apps: []WorkflowApp{
	//			},
	//		}
	//	}
	//},

	// This is not fast, but ok with just a few hundred thousand iterations :>
	log.Printf("[INFO] Starting mapping of labels from all %d apps", len(newapps))
	for categoryIndex, _ := range categories {
		categories[categoryIndex].AppLabels = []AppLabel{}
	}
	/*
			for labelIndex, label := range category.ActionLabels {
				//categories[categoryIndex].LabelApps = append(categories[categoryIndex].LabelApps, AppCategoryLabel{
				//	Label:          label,
				//	FormattedLabel: strings.ReplaceAll(strings.ToLower(label), " ", "_"),
				//	Apps:           []WorkflowApp{},
				//})
			}
		}
	*/

	for _, app := range newapps {
		if len(app.Name) == 0 {
			continue
		}

		if len(app.Categories) == 0 {
			//log.Printf("[INFO] No categories: %#v (%s)", app.Name, app.ID)
			continue
		}

		appLabels := []string{}
		for _, action := range app.Actions {
			// Compare with formatted label
			if len(action.CategoryLabel) > 0 {
				//&& strings.ReplaceAll(strings.ToLower(action.CategoryLabel[0]), " ", "_") == categories[categoryIndex].LabelApps[labelIndex].FormattedLabel {
				//log.Printf("[INFO] Found label %s in category %d with label %d. App: %s", label, categoryIndex, labelIndex, app.Name)
				appLabels = append(appLabels, action.CategoryLabel[0])

				//AppLabels    []AppCategoryLabel `json:"app_labels"`

			}
		}

		if len(appLabels) > 0 {
			log.Printf("[DEBUG] '%s' Got labels (%s): %#v", app.Name, app.Categories[0], appLabels)
			for categoryIndex, category := range categories {
				if strings.ToLower(category.Name) == strings.ToLower(app.Categories[0]) {
					newApp := AppLabel{
						AppName:    app.Name,
						LargeImage: app.LargeImage,
						ID:         app.ID,
					}

					// FIXME: May need to set the label to be the correct name according to the category's label
					for _, appLabel := range appLabels {
						newApp.Labels = append(newApp.Labels, LabelStruct{
							Category: category.Name,
							Label:    appLabel,
						})
					}

					categories[categoryIndex].AppLabels = append(categories[categoryIndex].AppLabels, newApp)
				}
			}
		}
	}

	newjson, err := json.Marshal(categories)
	if err != nil {
		log.Printf("[WARNING] Failed marshal in get categories: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Failed unpacking categories. Please try again."}`)))
		return
	}

	resp.WriteHeader(200)
	resp.Write(newjson)

}

func HandleRecommendationAction(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[AUDIT] Api authentication failed in modify recommendation: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(resp, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var recommendation RecommendationAction
	err = json.Unmarshal(body, &recommendation)
	if err != nil {
		log.Printf("[WARNING] Failed unmarshalling recommendation: %s", err)
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	availableActions := []string{"dismiss"}
	if !ArrayContains(availableActions, recommendation.Action) {
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)
	org, err := GetOrg(ctx, user.ActiveOrg.Id)
	if err != nil {
		log.Printf("[WARNING] Failed getting org '%s': %s", user.ActiveOrg.Id, err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Failed getting your org details"}`))
		return
	}

	changed := false
	for prioIndex, prio := range org.Priorities {
		if !prio.Active {
			continue
		}

		if prio.Name != recommendation.Name {
			//log.Printf("[DEBUG] '%s' is not '%s'", prio.Name, recommendation.Name)
			continue
		}

		// dismiss first, other later :)
		if recommendation.Action == "dismiss" {
			org.Priorities[prioIndex].Active = false
			changed = true
			break
		}
	}

	if changed {
		err = SetOrg(ctx, *org, org.Id)
		if err != nil {
			log.Printf("[WARNING] Failed updating org during priority updates: %s", err)
			resp.WriteHeader(500)
			resp.Write([]byte(`{"success": true}`))
			return
		}
	}

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

// Hard-coded to test out how we can generate next steps in workflows
// This could actually work when mapped back to usecases & with LLMs

// Mainly tested with Outlook Office365 for now
// Should be made based on:
// - Usecases and their structure
// - Active Apps (framework~)
// - LLMs
func HandleActionRecommendation(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[AUDIT] Api authentication failed in get action recommendations: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	_ = user

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(resp, "Error reading request body", http.StatusInternalServerError)
		return
	}
	var workflow Workflow

	workflowerr := json.Unmarshal(body, &workflow)
	if workflowerr != nil {
		log.Printf("[WARNING] Failed unmarshalling workflow: %s", workflowerr)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var recommendAction ActionRecommendations
	if len(workflow.Actions) > 0 {

		for key, _ := range workflow.Actions {
			var recommendations []Recommendations
			var action RecommendAction

			app := workflow.Actions[key].AppName + "_" + workflow.Actions[key].AppVersion
			if workflow.Actions[key].AppName == "Outlook_Office365" {
				if workflow.Actions[key].Name == "get_emails" {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID

					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "mark_email_as_read", AppVersion: "1.1.0", AppId: "accdaaf2eeba6a6ed43b2efc0112032d"})
					recommendations = append(recommendations, Recommendations{AppName: "Shuffle Tools", AppAction: "filter_list", AppVersion: "1.2.0", AppId: "3e2bdf9d5069fe3f4746c29d68785a6a"})
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "get_list_attachments", AppVersion: "1.1.0", AppId: "accdaaf2eeba6a6ed43b2efc0112032d"})
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "get_raw_email_as_file", AppVersion: "1.1.0", AppId: "accdaaf2eeba6a6ed43b2efc0112032d"})

					action.Recommendations = recommendations
					recommendAction.Actions = append(recommendAction.Actions, action)

				} else if workflow.Actions[key].Name == "get_raw_email_as_file" {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendations = append(recommendations, Recommendations{AppName: "email", AppAction: "parse_email_file"})
					action.Recommendations = recommendations
					recommendAction.Actions = append(recommendAction.Actions, action)
				} else {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendAction.Actions = append(recommendAction.Actions, action)
				}
			} else if workflow.Actions[key].AppName == "outlook-exchange" {
				if workflow.Actions[key].Name == "get_emails" {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "mark_email_as_read"})
					recommendations = append(recommendations, Recommendations{AppName: "Shuffle Tools", AppAction: "filter_list"})
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "move_email"})
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "delete_email"})
					action.Recommendations = recommendations
					recommendAction.Actions = append(recommendAction.Actions, action)
				} else {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendAction.Actions = append(recommendAction.Actions, action)
				}
			} else if workflow.Actions[key].AppName == "Gmail" {
				if workflow.Actions[key].Name == "get_gmail_users_messages_list" {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendations = append(recommendations, Recommendations{AppName: "Shuffle Tools", AppAction: "filter_list"})
					action.Recommendations = recommendations
					recommendAction.Actions = append(recommendAction.Actions, action)
				} else {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendAction.Actions = append(recommendAction.Actions, action)
				}
			} else if workflow.Actions[key].AppName == "TheHiveOpenAPI" {
				if workflow.Actions[key].Name == "get_list_alerts" {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "get_an_alert"})
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "patch_edit_an_alert"})
					action.Recommendations = recommendations
					recommendAction.Actions = append(recommendAction.Actions, action)
				} else {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendAction.Actions = append(recommendAction.Actions, action)
				}

			} else if workflow.Actions[key].AppName == "thehive" {
				if workflow.Actions[key].Name == "search_alert_title" {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "update_alert"})
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "close_alert"})
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "add_alert_artifact"})
					action.Recommendations = recommendations
					recommendAction.Actions = append(recommendAction.Actions, action)
				} else if workflow.Actions[key].Name == "create_case" {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "add_case_artifact"})
					recommendations = append(recommendations, Recommendations{AppName: workflow.Actions[key].AppName, AppAction: "update_case"})
					action.Recommendations = recommendations
					recommendAction.Actions = append(recommendAction.Actions, action)
				} else {
					action.AppName = app
					action.ActionId = workflow.Actions[key].ID
					recommendAction.Actions = append(recommendAction.Actions, action)
				}
			} else {
				action.AppName = app
				action.ActionId = workflow.Actions[key].ID
				recommendAction.Actions = append(recommendAction.Actions, action)
			}

		}
	}

	recommendAction.Success = true
	newjson, err := json.Marshal(recommendAction)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal recommendedAction: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write(newjson)

}

func HandleGetenvStats(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in stop executions: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")
	var fileId string
	if location[1] == "api" {
		if len(location) <= 4 {
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[4]
	}

	if user.Role != "admin" {
		log.Printf("[AUDIT] User isn't admin during stop executions")
		resp.WriteHeader(409)
		resp.Write([]byte(fmt.Sprintf(`{"success": false, "reason": "Must be admin to perform this action"}`)))
		return
	}

	ctx := GetContext(request)
	environmentName := fileId
	if len(fileId) != 36 {
		log.Printf("[DEBUG] Environment length %d for %s is not good for env Stats. Attempting to find the actual ID for it", len(fileId), fileId)

		environments, err := GetEnvironments(ctx, user.ActiveOrg.Id)
		if err != nil {
			log.Printf("[WARNING] Failed getting environments to validate: %s", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed to validate environment"}`))
			return
		}

		for _, environment := range environments {
			if environment.Name == fileId && len(environment.Id) > 0 {
				environmentName = fileId
				fileId = environment.Id
				break
			}
		}

		if len(fileId) != 36 {
			log.Printf("[WARNING] Failed getting environments to validate. New FileId: %s", fileId)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false, "reason": "Failed updating environment"}`))
			return
		}
	}

	// Should get stats for this
	_ = environmentName

	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(`{"success": true}`)))
}

func GetWorkflowSuggestions(ctx context.Context, user User, org *Org, orgUpdated bool) (*Org, bool) {
	// Loop workflows
	// Find "next" workflow to build
	// Find "untagged" workflows and map them to a usecase

	// 1. Suggest based on usecases
	// 2. Suggest public workflows (cloud)
	// 3. Use workflow template (local)
	var updated bool
	workflows, err := GetAllWorkflowsByQuery(ctx, user)
	if err != nil {
		log.Printf("[WARNING] No workflows found for user %s (2)", user.Id)
		return org, orgUpdated
	}

	log.Printf("[INFO] Finding workflow suggestions for %s (%s) based on %d workflows", org.Name, org.Id, len(workflows))
	for _, workflow := range workflows {
		for _, action := range workflow.Actions {
			if len(action.Category) == 0 {
				continue
			}

			//log.Printf("%s:%s = %s", action.AppName, action.AppVersion, action.Category)
			if org.SecurityFramework.Communication.Name == "" && (action.Category == "Communication" || action.Category == "email") {
				orgUpdated = true
				org.SecurityFramework.Communication = Category{
					Name:        action.Name,
					Count:       1,
					Description: "",
					LargeImage:  action.LargeImage,
					ID:          action.AppID,
				}
			}

			if org.SecurityFramework.Intel.Name == "" && action.Category == "Intel" {
				orgUpdated = true
				org.SecurityFramework.Intel = Category{
					Name:        action.Name,
					Count:       1,
					Description: "",
					LargeImage:  action.LargeImage,
					ID:          action.AppID,
				}
			}

			if org.SecurityFramework.Network.Name == "" && action.Category == "Network" {
				orgUpdated = true
				org.SecurityFramework.Network = Category{
					Name:        action.Name,
					Count:       1,
					Description: "",
					LargeImage:  action.LargeImage,
					ID:          action.AppID,
				}
			}

			if org.SecurityFramework.Assets.Name == "" && action.Category == "Assets" {
				orgUpdated = true
				org.SecurityFramework.Assets = Category{
					Name:        action.Name,
					Count:       1,
					Description: "",
					LargeImage:  action.LargeImage,
					ID:          action.AppID,
				}
			}

			if org.SecurityFramework.Cases.Name == "" && action.Category == "Cases" {
				orgUpdated = true
				org.SecurityFramework.Cases = Category{
					Name:        action.Name,
					Count:       1,
					Description: "",
					LargeImage:  action.LargeImage,
					ID:          action.AppID,
				}
			}

			if org.SecurityFramework.SIEM.Name == "" && action.Category == "SIEM" {
				orgUpdated = true
				org.SecurityFramework.SIEM = Category{
					Name:        action.Name,
					Count:       1,
					Description: "",
					LargeImage:  action.LargeImage,
					ID:          action.AppID,
				}
			}

			if org.SecurityFramework.EDR.Name == "" && action.Category == "EDR" {
				orgUpdated = true
				org.SecurityFramework.EDR = Category{
					Name:        action.Name,
					Count:       1,
					Description: "",
					LargeImage:  action.LargeImage,
					ID:          action.AppID,
				}
			}

			if org.SecurityFramework.IAM.Name == "" && action.Category == "IAM" {
				orgUpdated = true
				org.SecurityFramework.IAM = Category{
					Name:        action.Name,
					Count:       1,
					Description: "",
					LargeImage:  action.LargeImage,
					ID:          action.AppID,
				}
			}
		}
	}

	// Checking again to see if specifying either should be a priority
	if org.SecurityFramework.SIEM.Name == "" || org.SecurityFramework.EDR.Name == "" || org.SecurityFramework.Communication.Name == "" {
		org, updated = AddPriority(*org, Priority{
			Name:        "Apps for Email, EDR & SIEM should be specified",
			Description: "The most common usecases are based on Email, EDR & SIEM. If these aren't specified Shuffle won't be used optimally.",
			Type:        "apps",
			Active:      true,
			URL:         fmt.Sprintf("/welcome?tab=2"),
			Severity:    2,
		}, updated)

		if updated {
			orgUpdated = true
		}
	}

	// Checking which workflows SHOULD have a usecase attached to them
	for _, workflow := range workflows {
		if len(workflow.UsecaseIds) != 0 {
			continue
		}

		//log.Printf("[INFO] No usecase for workflow %s", workflow.Name)

		// Sample: If email (get/trigger) & cases (create ticket) in same workflow -> email usecase = done
		// If excel/sheets is used, reporting
		// Add keywords to usecases? Check if anything matching in:
		// - name
		// - action name
		// - action label(s)
		// - action description
	}

	// Matching org priority with usecases & previously built workflows
	var usecases UsecaseLinks
	err = json.Unmarshal([]byte(GetUsecaseData()), &usecases)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal usecase data (priorities): %s", err)
	} else {
		log.Printf("[DEBUG] Got parsed usecases for %s - should check priority vs mainpriority (%s)", org.Name, org.MainPriority)

		selectedAppName := ""
		selectedAppImage := ""
		innerUpdate := false
		for usecaseIndex, usecase := range usecases {
			//if usecase.Name != org.MainPriority {
			//	continue
			//}

			// match them with usecases here
			for _, workflow := range workflows {
				if len(workflow.UsecaseIds) == 0 {
					continue
				}

				// Fidning matching usecase for workflow
				for _, workflowUsecase := range workflow.UsecaseIds {
					newUsecasename := strings.ToLower(workflowUsecase)

					for subusecaseIndex, subusecase := range usecase.List {
						if newUsecasename == strings.ToLower(subusecase.Name) {
							usecases[usecaseIndex].List[subusecaseIndex].Matches = append(usecases[usecaseIndex].List[subusecaseIndex].Matches, workflow)
							break
						}
					}
				}
			}

			// Sort sub-usecases by priority
			slice.Sort(usecase.List[:], func(i, j int) bool {
				return usecase.List[i].Priority > usecase.List[j].Priority
			})

			//log.Printf("[DEBUG] Priorities for %s", usecase.Name)
			cntAdded := 0
			for _, subusecase := range usecase.List {
				// Check if it has a workflow attached to it too?
				//log.Printf("%s = %d. Matches: %d", subusecase.Name, subusecase.Priority, len(subusecase.Matches))

				// Matches = matching usecases that have workflows attached to them
				// This means if it exists, don't add it as priority
				if len(subusecase.Matches) > 0 {
					continue
				}

				if strings.ToLower(subusecase.Type) == "iam" {
					if org.SecurityFramework.IAM.Name == "" {
						continue
					}

					selectedAppName = org.SecurityFramework.IAM.Name
					selectedAppImage = org.SecurityFramework.IAM.LargeImage
				}

				if strings.ToLower(subusecase.Type) == "siem" {
					if org.SecurityFramework.SIEM.Name == "" {
						continue
					}

					selectedAppName = org.SecurityFramework.SIEM.Name
					selectedAppImage = org.SecurityFramework.SIEM.LargeImage
				}

				if strings.ToLower(subusecase.Type) == "edr" {
					if org.SecurityFramework.EDR.Name == "" {
						continue
					}

					selectedAppName = org.SecurityFramework.EDR.Name
					selectedAppImage = org.SecurityFramework.EDR.LargeImage
				}

				if strings.ToLower(subusecase.Type) == "communication" {
					if org.SecurityFramework.Communication.Name == "" {
						continue
					}

					selectedAppName = org.SecurityFramework.Communication.Name
					selectedAppImage = org.SecurityFramework.Communication.LargeImage
				}

				if strings.ToLower(subusecase.Type) == "assets" {
					if org.SecurityFramework.Assets.Name == "" {
						continue
					}

					selectedAppName = org.SecurityFramework.Assets.Name
					selectedAppImage = org.SecurityFramework.Assets.LargeImage
				}

				if strings.ToLower(subusecase.Type) == "cases" {
					if org.SecurityFramework.Cases.Name == "" {
						continue
					}

					selectedAppName = org.SecurityFramework.Cases.Name
					selectedAppImage = org.SecurityFramework.Cases.LargeImage
				}

				if strings.ToLower(subusecase.Type) == "network" {
					if org.SecurityFramework.Network.Name == "" {
						continue
					}

					selectedAppName = org.SecurityFramework.Network.Name
					selectedAppImage = org.SecurityFramework.Network.LargeImage
				}

				if strings.ToLower(subusecase.Type) == "intel" {
					if org.SecurityFramework.Intel.Name == "" {
						continue
					}

					selectedAppName = org.SecurityFramework.Intel.Name
					selectedAppImage = org.SecurityFramework.Intel.LargeImage
				}

				usecaseDescription := "A priority usecase for your organization."
				if len(selectedAppName) > 0 && len(selectedAppImage) > 0 {
					usecaseDescription = fmt.Sprintf("%s&%s", strings.Replace(selectedAppName, "_", " ", -1), selectedAppImage)

					// Adding "last" node as well
					if strings.ToLower(subusecase.Last) == "iam" && org.SecurityFramework.IAM.Name != "" && org.SecurityFramework.IAM.LargeImage != "" {
						usecaseDescription = fmt.Sprintf("%s&%s&%s", usecaseDescription, strings.Replace(org.SecurityFramework.IAM.Name, "_", " ", -1), org.SecurityFramework.IAM.LargeImage)
					} else if strings.ToLower(subusecase.Last) == "siem" && org.SecurityFramework.SIEM.Name != "" && org.SecurityFramework.SIEM.LargeImage != "" {
						usecaseDescription = fmt.Sprintf("%s&%s&%s", usecaseDescription, strings.Replace(org.SecurityFramework.SIEM.Name, "_", " ", -1), org.SecurityFramework.SIEM.LargeImage)
					} else if strings.ToLower(subusecase.Last) == "edr" && org.SecurityFramework.EDR.Name != "" && org.SecurityFramework.EDR.LargeImage != "" {
						usecaseDescription = fmt.Sprintf("%s&%s&%s", usecaseDescription, strings.Replace(org.SecurityFramework.EDR.Name, "_", " ", -1), org.SecurityFramework.EDR.LargeImage)
					} else if strings.ToLower(subusecase.Last) == "communication" && org.SecurityFramework.Communication.Name != "" && org.SecurityFramework.Communication.LargeImage != "" {
						usecaseDescription = fmt.Sprintf("%s&%s&%s", usecaseDescription, strings.Replace(org.SecurityFramework.Communication.Name, "_", " ", -1), org.SecurityFramework.Communication.LargeImage)
					} else if strings.ToLower(subusecase.Last) == "assets" && org.SecurityFramework.Assets.Name != "" && org.SecurityFramework.Assets.LargeImage != "" {
						usecaseDescription = fmt.Sprintf("%s&%s&%s", usecaseDescription, strings.Replace(org.SecurityFramework.Assets.Name, "_", " ", -1), org.SecurityFramework.Assets.LargeImage)
					} else if strings.ToLower(subusecase.Last) == "cases" && org.SecurityFramework.Cases.Name != "" && org.SecurityFramework.Cases.LargeImage != "" {
						usecaseDescription = fmt.Sprintf("%s&%s&%s", usecaseDescription, strings.Replace(org.SecurityFramework.Cases.Name, "_", " ", -1), org.SecurityFramework.Cases.LargeImage)
					} else if strings.ToLower(subusecase.Last) == "network" && org.SecurityFramework.Network.Name != "" && org.SecurityFramework.Network.LargeImage != "" {
						usecaseDescription = fmt.Sprintf("%s&%s&%s", usecaseDescription, strings.Replace(org.SecurityFramework.Network.Name, "_", " ", -1), org.SecurityFramework.Network.LargeImage)
					} else if strings.ToLower(subusecase.Last) == "intel" && org.SecurityFramework.Intel.Name != "" && org.SecurityFramework.Intel.LargeImage != "" {
						usecaseDescription = fmt.Sprintf("%s&%s&%s", usecaseDescription, strings.Replace(org.SecurityFramework.Intel.Name, "_", " ", -1), org.SecurityFramework.Intel.LargeImage)
					}
				}

				// Should find info about the usecase
				// No description as this has custom rendering
				org, innerUpdate = AddPriority(*org, Priority{
					Name:        fmt.Sprintf("Suggested Usecase: %s", subusecase.Name),
					Description: usecaseDescription,
					Type:        "usecase",
					Active:      true,
					URL:         fmt.Sprintf("/usecases?selected_object=%s", subusecase.Name),
					Severity:    3,
				}, updated)

				if innerUpdate {
					log.Printf("[DEBUG] Added priority for %s", subusecase.Name)

					cntAdded += 1
					orgUpdated = true
					break
				}
			}

			if innerUpdate {
				break
			}
		}
	}

	return org, orgUpdated
}
