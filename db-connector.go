package shuffle

import (
	"cloud.google.com/go/datastore"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"google.golang.org/api/iterator"
	"google.golang.org/appengine/memcache"
)

var err error
var requestCache *cache.Cache

var maxCacheSize = 1020000

//var maxCacheSize = 2000000

// Cache handlers
func DeleteCache(ctx context.Context, name string) error {
	if project.Environment == "cloud" {
		return memcache.Delete(ctx, name)
	} else if project.Environment == "onprem" {
		requestCache.Delete(name)
		return nil
	} else {
		return errors.New(fmt.Sprintf("No cache handler for environment %s yet WHILE DELETING", project.Environment))
	}

	return errors.New(fmt.Sprintf("No cache found for %s when DELETING cache", name))
}

// Cache handlers
func GetCache(ctx context.Context, name string) (interface{}, error) {
	if project.Environment == "cloud" {
		if item, err := memcache.Get(ctx, name); err == memcache.ErrCacheMiss {
		} else if err != nil {
			return "", errors.New(fmt.Sprintf("Failed getting CLOUD cache for %s: %s", name, err))
		} else {
			// Loops if cachesize is more than max allowed in memcache (multikey)
			if len(item.Value) == maxCacheSize {
				totalData := item.Value
				keyCount := 1
				keyname := fmt.Sprintf("%s_%d", name, keyCount)
				for {
					if item, err := memcache.Get(ctx, keyname); err == memcache.ErrCacheMiss {
						break
					} else {
						totalData = append(totalData, item.Value...)

						//log.Printf("%d - %d = ", len(item.Value), maxCacheSize)
						if len(item.Value) != maxCacheSize {
							break
						}
					}

					keyCount += 1
					keyname = fmt.Sprintf("%s_%d", name, keyCount)
				}

				log.Printf("[INFO] CACHE: TOTAL SIZE FOR %s: %d", name, len(totalData))
				return totalData, nil
			} else {
				return item.Value, nil
			}
		}
	} else if project.Environment == "onprem" {
		//log.Printf("[INFO] GETTING CACHE FOR %s ONPREM", name)
		if value, found := requestCache.Get(name); found {
			return value, nil
		} else {
			return "", errors.New(fmt.Sprintf("Failed getting ONPREM cache for %s", name))
		}
	} else {
		return "", errors.New(fmt.Sprintf("No cache handler for environment %s yet", project.Environment))
	}

	return "", errors.New(fmt.Sprintf("No cache found for %s", name))
}

func SetCache(ctx context.Context, name string, data []byte) error {
	//log.Printf("DATA SIZE: %d", len(data))
	// Maxsize ish~

	if project.Environment == "cloud" {
		if len(data) > maxCacheSize*10 {
			return errors.New(fmt.Sprintf("Couldn't set cache for %s - too large: %d > %d", name, len(data), maxCacheSize*10))
		}
		loop := false
		if len(data) > maxCacheSize {
			loop = true
			//log.Printf("Should make multiple cache items for %s", name)
		}

		// Custom for larger sizes. Max is maxSize*10 when being set
		if loop {
			currentChunk := 0
			keyAmount := 0
			totalAdded := 0
			chunkSize := maxCacheSize
			nextStep := chunkSize
			keyname := name

			for {
				if len(data) < nextStep {
					nextStep = len(data)
				}

				//log.Printf("%d - %d = ", currentChunk, nextStep)
				parsedData := data[currentChunk:nextStep]
				item := &memcache.Item{
					Key:        keyname,
					Value:      parsedData,
					Expiration: time.Minute * 30,
				}

				if err := memcache.Set(ctx, item); err != nil {
					log.Printf("[WARNING] Failed setting cache for %s: %s", keyname, err)
					break
				} else {
					totalAdded += chunkSize
					currentChunk = nextStep
					nextStep += chunkSize

					keyAmount += 1
					//log.Printf("%s: %d: %d", keyname, totalAdded, len(data))

					keyname = fmt.Sprintf("%s_%d", name, keyAmount)
					if totalAdded > len(data) {
						break
					}
				}
			}

			log.Printf("[INFO] Set app cache with length %d and %d keys", len(data), keyAmount)
		} else {
			item := &memcache.Item{
				Key:        name,
				Value:      data,
				Expiration: time.Minute * 30,
			}

			if err := memcache.Set(ctx, item); err != nil {
				log.Printf("[WARNING] Failed setting cache for %s: %s", name, err)
			}
		}

		return nil
	} else if project.Environment == "onprem" {
		//log.Printf("SETTING CACHE FOR %s ONPREM", name)
		requestCache.Set(name, data, cache.DefaultExpiration)
	} else {
		return errors.New(fmt.Sprintf("No cache handler for environment %s yet", project.Environment))
	}

	return nil
}

func GetDatastoreClient(ctx context.Context, projectID string) (datastore.Client, error) {
	// FIXME - this doesn't work
	//client, err := datastore.NewClient(ctx, projectID, option.WithCredentialsFile(test"))
	client, err := datastore.NewClient(ctx, projectID)
	//client, err := datastore.NewClient(ctx, projectID, option.WithCredentialsFile("test"))
	if err != nil {
		return datastore.Client{}, err
	}

	return *client, nil
}

func SetWorkflowAppDatastore(ctx context.Context, workflowapp WorkflowApp, id string) error {
	key := datastore.NameKey("workflowapp", id, nil)

	// New struct, to not add body, author etc
	if _, err := project.Dbclient.Put(ctx, key, &workflowapp); err != nil {
		log.Printf("[WARNING] Error adding workflow app: %s", err)
		return err
	}

	return nil
}

func SetWorkflowExecution(ctx context.Context, workflowExecution WorkflowExecution, dbSave bool) error {
	//log.Printf("\n\n\nRESULT: %s\n\n\n", workflowExecution.Status)
	if len(workflowExecution.ExecutionId) == 0 {
		log.Printf("[WARNING] Workflowexeciton executionId can't be empty.")
		return errors.New("ExecutionId can't be empty.")
	}

	cacheKey := fmt.Sprintf("workflowexecution-%s", workflowExecution.ExecutionId)
	executionData, err := json.Marshal(workflowExecution)
	if err != nil {
		log.Printf("[WARNING] Failed marshalling execution for cache: %s", err)

		err = SetCache(ctx, cacheKey, executionData)
		if err != nil {
			log.Printf("[WARNING] Failed updating execution cache: %s", err)
		}
	} else {
		log.Printf("[INFO] Set execution cache for workflowexecution %s", cacheKey)
	}

	//requestCache.Set(cacheKey, &workflowExecution, cache.DefaultExpiration)
	if !dbSave && workflowExecution.Status == "EXECUTING" && len(workflowExecution.Results) > 1 {
		//log.Printf("[WARNING] SHOULD skip DB saving for execution")
		return nil
	}

	// New struct, to not add body, author etc
	key := datastore.NameKey("workflowexecution", workflowExecution.ExecutionId, nil)
	if _, err := project.Dbclient.Put(ctx, key, &workflowExecution); err != nil {
		log.Printf("Error adding workflow_execution: %s", err)
		return err
	}

	return nil
}

type ExecutionVariableWrapper struct {
	StartNode    string              `json:"startnode"`
	Children     map[string][]string `json:"children"`
	Parents      map[string][]string `json:"parents""`
	Visited      []string            `json:"visited"`
	Executed     []string            `json:"executed"`
	NextActions  []string            `json:"nextActions"`
	Environments []string            `json:"environments"`
	Extra        int                 `json:"extra"`
}

// Initializes an execution's extra variables
func SetInitExecutionVariables(ctx context.Context, workflowExecution WorkflowExecution) {
	environments := []string{}
	nextActions := []string{}
	startAction := ""
	extra := 0
	parents := map[string][]string{}
	children := map[string][]string{}

	// Hmm
	triggersHandled := []string{}

	for _, action := range workflowExecution.Workflow.Actions {
		if !ArrayContains(environments, action.Environment) {
			environments = append(environments, action.Environment)
		}

		if action.ID == workflowExecution.Start {
			/*
				functionName = fmt.Sprintf("%s-%s", action.AppName, action.AppVersion)

				if !action.Sharing {
					functionName = fmt.Sprintf("%s-%s", action.AppName, action.PrivateID)
				}
			*/

			startAction = action.ID
		}
	}

	nextActions = append(nextActions, startAction)
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

		for _, trigger := range workflowExecution.Workflow.Triggers {
			//log.Printf("Appname trigger (0): %s", trigger.AppName)
			if trigger.AppName == "User Input" || trigger.AppName == "Shuffle Workflow" {
				//log.Printf("%s is a special trigger. Checking where.", trigger.AppName)

				found := false
				for _, check := range triggersHandled {
					if check == trigger.ID {
						found = true
						break
					}
				}

				if !found {
					extra += 1
				} else {
					triggersHandled = append(triggersHandled, trigger.ID)
				}

				if trigger.ID == branch.SourceID {
					log.Printf("Trigger %s is the source!", trigger.AppName)
					sourceFound = true
				} else if trigger.ID == branch.DestinationID {
					log.Printf("Trigger %s is the destination!", trigger.AppName)
					destinationFound = true
				}
			}
		}

		if sourceFound {
			parents[branch.DestinationID] = append(parents[branch.DestinationID], branch.SourceID)
		} else {
			log.Printf("[INFO] ID %s was not found in actions! Skipping parent. (TRIGGER?)", branch.SourceID)
		}

		if destinationFound {
			children[branch.SourceID] = append(children[branch.SourceID], branch.DestinationID)
		} else {
			log.Printf("[INFO] ID %s was not found in actions! Skipping child. (TRIGGER?)", branch.SourceID)
		}
	}

	/*
		log.Printf("\n\nEnvironments: %#v", environments)
		log.Printf("Startnode: %s", startAction)
		log.Printf("Parents: %#v", parents)
		log.Printf("NextActions: %#v", nextActions)
		log.Printf("Extra: %d", extra)
		log.Printf("Children: %s", children)
	*/

	UpdateExecutionVariables(ctx, workflowExecution.ExecutionId, startAction, children, parents, []string{startAction}, []string{startAction}, nextActions, environments, extra)

}

func UpdateExecutionVariables(ctx context.Context, executionId, startnode string, children, parents map[string][]string, visited, executed, nextActions, environments []string, extra int) {
	cacheKey := fmt.Sprintf("%s-actions", executionId)
	//log.Printf("\n\nSHOULD UPDATE VARIABLES FOR %s\n\n", executionId)
	_ = cacheKey

	newVariableWrapper := ExecutionVariableWrapper{
		StartNode:    startnode,
		Children:     children,
		Parents:      parents,
		NextActions:  nextActions,
		Environments: environments,
		Extra:        extra,
		Visited:      visited,
		Executed:     visited,
	}

	variableWrapperData, err := json.Marshal(newVariableWrapper)
	if err != nil {
		log.Printf("[WARNING] Failed marshalling execution: %s", err)
		return
	}

	err = SetCache(ctx, cacheKey, variableWrapperData)
	if err != nil {
		log.Printf("[WARNING] Failed updating execution: %s", err)
	}

	log.Printf("[INFO] Successfully set cache for execution variables %s\n\n", cacheKey)
}

func GetExecutionVariables(ctx context.Context, executionId string) (string, int, map[string][]string, map[string][]string, []string, []string, []string, []string) {

	cacheKey := fmt.Sprintf("%s-actions", executionId)
	wrapper := &ExecutionVariableWrapper{}
	if project.CacheDb {
		cache, err := GetCache(ctx, cacheKey)
		if err == nil {
			cacheData := []byte(cache.([]uint8))
			//log.Printf("CACHEDATA: %#v", cacheData)
			err = json.Unmarshal(cacheData, &wrapper)
			if err == nil {
				return wrapper.StartNode, wrapper.Extra, wrapper.Children, wrapper.Parents, wrapper.Visited, wrapper.Executed, wrapper.NextActions, wrapper.Environments
			}
		} else {
			//log.Printf("[INFO] Failed getting cache for execution variables data %s: %s", executionId, err)
		}
	}

	return "", 0, map[string][]string{}, map[string][]string{}, []string{}, []string{}, []string{}, []string{}
}

func GetWorkflowExecution(ctx context.Context, id string) (*WorkflowExecution, error) {
	cacheKey := fmt.Sprintf("workflowexecution-%s", id)
	workflowExecution := &WorkflowExecution{}
	if project.CacheDb {
		cache, err := GetCache(ctx, cacheKey)
		if err == nil {
			cacheData := []byte(cache.([]uint8))
			//log.Printf("CACHEDATA: %#v", cacheData)
			err = json.Unmarshal(cacheData, &workflowExecution)
			if err == nil {
				return workflowExecution, nil
			}
		} else {
			//log.Printf("[INFO] Failed getting cache for workflow execution: %s", err)
		}
	}

	key := datastore.NameKey("workflowexecution", strings.ToLower(id), nil)
	if err := project.Dbclient.Get(ctx, key, workflowExecution); err != nil {
		return &WorkflowExecution{}, err
	}

	if project.CacheDb {
		newexecution, err := json.Marshal(workflowExecution)
		if err != nil {
			log.Printf("[WARNING] Failed marshalling execution: %s", err)
			return workflowExecution, nil
		}

		err = SetCache(ctx, id, newexecution)
		if err != nil {
			log.Printf("[WARNING] Failed updating execution: %s", err)
		}
	}

	return workflowExecution, nil
}

func GetApp(ctx context.Context, id string, user User) (*WorkflowApp, error) {
	key := datastore.NameKey("workflowapp", strings.ToLower(id), nil)
	workflowApp := &WorkflowApp{}
	if err := project.Dbclient.Get(ctx, key, workflowApp); err != nil {
		for _, app := range user.PrivateApps {
			if app.ID == id {
				return &app, nil
			}
		}

		return &WorkflowApp{}, err
	}

	return workflowApp, nil
}

func GetWorkflow(ctx context.Context, id string) (*Workflow, error) {

	key := datastore.NameKey("workflow", strings.ToLower(id), nil)
	workflow := &Workflow{}
	if err := project.Dbclient.Get(ctx, key, workflow); err != nil {
		return &Workflow{}, err
	}

	return workflow, nil
}

func GetAllWorkflows(ctx context.Context, orgId string) ([]Workflow, error) {
	var allworkflows []Workflow
	q := datastore.NewQuery("workflow").Filter("org_id = ", orgId)

	_, err := project.Dbclient.GetAll(ctx, q, &allworkflows)
	if err != nil {
		return []Workflow{}, err
	}

	return allworkflows, nil
}

// ListBooks returns a list of books, ordered by title.
func GetOrg(ctx context.Context, id string) (*Org, error) {
	curOrg := &Org{}
	if project.CacheDb {
		cache, err := GetCache(ctx, id)
		if err == nil {
			cacheData := []byte(cache.([]uint8))
			//log.Printf("CACHEDATA: %#v", cacheData)
			err = json.Unmarshal(cacheData, &curOrg)
			if err == nil {
				return curOrg, nil
			}
		} else {
			//log.Printf("[INFO] Failed getting cache for org: %s", err)
		}
	}

	key := datastore.NameKey("Organizations", id, nil)
	if err := project.Dbclient.Get(ctx, key, curOrg); err != nil {
		return &Org{}, err
	}

	for _, user := range curOrg.Users {
		user.Password = ""
		user.Session = ""
		user.ResetReference = ""
		user.PrivateApps = []WorkflowApp{}
		user.VerificationToken = ""
		user.ApiKey = ""
		user.Executions = ExecutionInfo{}
	}

	if project.CacheDb {
		neworg, err := json.Marshal(curOrg)
		if err != nil {
			log.Printf("[WARNING] Failed marshalling org: %s", err)
			return curOrg, nil
		}

		err = SetCache(ctx, id, neworg)
		if err != nil {
			log.Printf("[WARNING] Failed updating cache: %s", err)
		}
	}

	return curOrg, nil
}

func SetOrg(ctx context.Context, data Org, id string) error {
	timeNow := int64(time.Now().Unix())
	if data.Created == 0 {
		data.Created = timeNow
	}

	data.Edited = timeNow

	// clear session_token and API_token for user
	k := datastore.NameKey("Organizations", id, nil)
	if _, err := project.Dbclient.Put(ctx, k, &data); err != nil {
		log.Println(err)
		return err
	}

	if project.CacheDb {
		neworg, err := json.Marshal(data)
		if err != nil {
			return nil
		}

		err = SetCache(ctx, id, neworg)
		if err != nil {
			log.Printf("Failed setting cache: %s", err)
			//DeleteCache(neworg)
		}
	}

	return nil
}

func GetSession(ctx context.Context, thissession string) (*Session, error) {
	session := &Session{}
	cache, err := GetCache(ctx, thissession)
	if err == nil {
		cacheData := []byte(cache.([]uint8))
		//log.Printf("CACHEDATA: %#v", cacheData)
		err = json.Unmarshal(cacheData, &session)
		if err == nil {
			return session, nil
		}
	} else {
		//log.Printf("[WARNING] Error getting session cache for %s: %v", thissession, err)
	}

	key := datastore.NameKey("sessions", thissession, nil)
	if err := project.Dbclient.Get(ctx, key, session); err != nil {
		return &Session{}, err
	}

	if project.CacheDb {
		data, err := json.Marshal(thissession)
		if err != nil {
			log.Printf("[WARNING] Failed marshalling session: %s", err)
			return session, nil
		}

		err = SetCache(ctx, thissession, data)
		if err != nil {
			log.Printf("[WARNING] Failed updating session cache: %s", err)
		}
	}

	return session, nil
}

// Index = Username
func DeleteKey(ctx context.Context, entity string, value string) error {
	// Non indexed User data
	key1 := datastore.NameKey(entity, value, nil)

	err = project.Dbclient.Delete(ctx, key1)
	if err != nil {
		log.Printf("Error deleting %s from %s: %s", value, entity, err)
		return err
	}

	return nil
}

// Index = Username
func SetApikey(ctx context.Context, Userdata User) error {

	// Non indexed User data
	newapiUser := new(Userapi)
	newapiUser.ApiKey = Userdata.ApiKey
	newapiUser.Username = Userdata.Username
	key1 := datastore.NameKey("apikey", newapiUser.ApiKey, nil)

	// New struct, to not add body, author etc
	if _, err := project.Dbclient.Put(ctx, key1, newapiUser); err != nil {
		log.Printf("Error adding apikey: %s", err)
		return err
	}

	return nil
}

func SetOpenApiDatastore(ctx context.Context, id string, data ParsedOpenApi) error {
	k := datastore.NameKey("openapi3", id, nil)
	if _, err := project.Dbclient.Put(ctx, k, &data); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetOpenApiDatastore(ctx context.Context, id string) (ParsedOpenApi, error) {
	key := datastore.NameKey("openapi3", id, nil)
	api := &ParsedOpenApi{}
	if err := project.Dbclient.Get(ctx, key, api); err != nil {
		return ParsedOpenApi{}, err
	}

	return *api, nil
}

// Index = Username
func SetSession(ctx context.Context, user User, value string) error {
	parsedKey := strings.ToLower(user.Username)
	if project.Environment != "cloud" {
		parsedKey = user.Id
	}

	// Non indexed User data
	user.Session = value
	key1 := datastore.NameKey("Users", parsedKey, nil)

	// New struct, to not add body, author etc
	if _, err := project.Dbclient.Put(ctx, key1, &user); err != nil {
		log.Printf("[WARNING] Error adding Usersession: %s", err)
		return err
	}

	if len(user.Session) > 0 {
		// Indexed session data
		sessiondata := new(Session)
		sessiondata.Username = user.Username
		sessiondata.Session = user.Session
		sessiondata.Id = user.Id
		key2 := datastore.NameKey("sessions", sessiondata.Session, nil)

		if _, err := project.Dbclient.Put(ctx, key2, sessiondata); err != nil {
			log.Printf("Error adding session: %s", err)
			return err
		}
	}

	return nil
}

// ListBooks returns a list of books, ordered by title.
func GetUser(ctx context.Context, username string) (*User, error) {
	curUser := &User{}

	parsedKey := strings.ToLower(username)
	cacheKey := fmt.Sprintf("user_%s", parsedKey)
	if project.CacheDb {
		cache, err := GetCache(ctx, cacheKey)
		if err == nil {
			cacheData := []byte(cache.([]uint8))
			err = json.Unmarshal(cacheData, &curUser)
			if err == nil {
				return curUser, nil
			}
		} else {
			//log.Printf("[INFO] Failed getting cache for user: %s", err)
		}
	}

	key := datastore.NameKey("Users", parsedKey, nil)
	if err := project.Dbclient.Get(ctx, key, curUser); err != nil {
		return &User{}, err
	}

	if project.CacheDb {
		data, err := json.Marshal(curUser)
		if err != nil {
			log.Printf("[WARNING] Failed marshalling user: %s", err)
			return curUser, nil
		}

		err = SetCache(ctx, cacheKey, data)
		if err != nil {
			log.Printf("[WARNING] Failed updating cache: %s", err)
		}
	}

	return curUser, nil
}

func SetUser(ctx context.Context, user *User) error {
	log.Printf("[INFO] Updating a user that has the role %s with %d apps", user.Role, len(user.PrivateApps))
	user = fixUserOrg(ctx, user)

	// clear session_token and API_token for user
	parsedKey := strings.ToLower(user.Username)
	if project.Environment != "cloud" {
		parsedKey = user.Id
	}

	k := datastore.NameKey("Users", parsedKey, nil)
	if _, err := project.Dbclient.Put(ctx, k, user); err != nil {
		log.Printf("[WARNING] Error updating user: %s", err)
		return err
	}

	DeleteCache(ctx, user.ApiKey)
	DeleteCache(ctx, user.Session)

	if project.CacheDb {
		cacheKey := fmt.Sprintf("user_%s", parsedKey)
		data, err := json.Marshal(user)
		if err != nil {
			log.Printf("[WARNING] Failed marshalling user: %s", err)
			return nil
		}

		err = SetCache(ctx, cacheKey, data)
		if err != nil {
			log.Printf("[WARNING] Failed updating user cache: %s", err)
		}
	}

	return nil
}

func getDatastoreClient(ctx context.Context, projectID string) (datastore.Client, error) {
	// FIXME - this doesn't work
	//client, err := datastore.NewClient(ctx, projectID, option.WithCredentialsFile(test"))
	client, err := datastore.NewClient(ctx, projectID)
	//client, err := datastore.NewClient(ctx, projectID, option.WithCredentialsFile("test"))
	if err != nil {
		return datastore.Client{}, err
	}

	return *client, nil
}

func fixUserOrg(ctx context.Context, user *User) *User {
	found := false
	for _, id := range user.Orgs {
		if user.ActiveOrg.Id == id {
			found = true
			break
		}
	}

	if !found {
		user.Orgs = append(user.Orgs, user.ActiveOrg.Id)
	}

	innerUser := *user
	innerUser.PrivateApps = []WorkflowApp{}
	innerUser.Executions = ExecutionInfo{}
	innerUser.Limits = UserLimits{}
	innerUser.Authentication = []UserAuth{}
	innerUser.Password = ""

	// Might be vulnerable to timing attacks.
	for _, orgId := range user.Orgs {
		if len(orgId) == 0 {
			continue
		}

		org, err := GetOrg(ctx, orgId)
		if err != nil {
			log.Printf("Error getting org %s", orgId)
			continue
		}

		orgIndex := 0
		userFound := false
		for index, orgUser := range org.Users {
			if orgUser.Id == user.Id {
				orgIndex = index
				userFound = true
				break
			}
		}

		if userFound {
			org.Users[orgIndex] = innerUser
		} else {
			org.Users = append(org.Users, innerUser)
		}

		err = SetOrg(ctx, *org, orgId)
		if err != nil {
			log.Printf("Failed setting org %s", orgId)
		}
	}

	return user
}

func GetAllWorkflowAppAuth(ctx context.Context, OrgId string) ([]AppAuthenticationStorage, error) {
	var allworkflowapps []AppAuthenticationStorage
	q := datastore.NewQuery("workflowappauth").Filter("org_id = ", OrgId)

	_, err = project.Dbclient.GetAll(ctx, q, &allworkflowapps)
	if err != nil {
		return []AppAuthenticationStorage{}, err
	}

	return allworkflowapps, nil
}

func GetEnvironments(ctx context.Context, OrgId string) ([]Environment, error) {
	var environments []Environment
	q := datastore.NewQuery("Environments").Filter("org_id =", OrgId)

	_, err = project.Dbclient.GetAll(ctx, q, &environments)
	if err != nil {
		return []Environment{}, err
	}

	return environments, nil
}

// Gets apps based on a new schema instead of looping everything
// Primarily made for cloud. Load in this order:
// 1. Get ORGs' private apps
// 2. Get USERs' private apps
// 3. Get PUBLIC apps
func GetPrioritizedApps(ctx context.Context, user User) ([]WorkflowApp, error) {
	allApps := []WorkflowApp{}
	//log.Printf("[INFO] LOOPING REAL APPS: %d. Private: %d", len(user.PrivateApps))

	cacheKey := fmt.Sprintf("apps_%s", user.Id)
	if project.CacheDb {
		cache, err := GetCache(ctx, cacheKey)
		if err == nil {
			cacheData := []byte(cache.([]uint8))
			err = json.Unmarshal(cacheData, &allApps)
			if err == nil {
				return allApps, nil
			} else {
				log.Println(string(cacheData))
				log.Printf("Failed unmarshaling apps: %s", err)
				log.Printf("DATALEN: %d", len(cacheData))
			}
		} else {
			log.Printf("[INFO] Failed getting cache for apps with KEY %s: %s", cacheKey, err)
		}
	}

	maxLen := 100
	cursorStr := ""
	limit := 100
	allApps = user.PrivateApps
	query := datastore.NewQuery("workflowapp").Filter("reference_org =", user.ActiveOrg.Id).Limit(limit)
	for {
		it := project.Dbclient.Run(ctx, query)

		for {
			innerApp := WorkflowApp{}
			_, err := it.Next(&innerApp)
			if err != nil {
				if strings.Contains(fmt.Sprintf("%s", err), "cannot load field") {
					log.Printf("[WARNING] Error in reference_org load: %s.", err)
					continue
				}

				log.Printf("[WARNING] No more apps for %s in org app load? Breaking: %s.", user.Username, err)
				break
			}

			found := false
			//log.Printf("ACTIONS: %d - %s", len(app.Actions), app.Name)
			for _, loopedApp := range allApps {
				if loopedApp.Name == innerApp.Name || loopedApp.ID == innerApp.ID {
					found = true
					break
				}
			}

			if !found {
				allApps = append(allApps, innerApp)
			}
		}

		if err != iterator.Done {
			//log.Printf("[INFO] Failed fetching results: %v", err)
			//break
		}

		// Get the cursor for the next page of results.
		nextCursor, err := it.Cursor()
		if err != nil {
			log.Printf("Cursorerror: %s", err)
			break
		} else {
			//log.Printf("NEXTCURSOR: %s", nextCursor)
			nextStr := fmt.Sprintf("%s", nextCursor)
			if cursorStr == nextStr {
				break
			}

			cursorStr = nextStr
			query = query.Start(nextCursor)
			//cursorStr = nextCursor
			//break
		}

		if len(allApps) > maxLen {
			break
		}
	}

	query = datastore.NewQuery("workflowapp").Filter("public =", true).Limit(limit)
	for {
		it := project.Dbclient.Run(ctx, query)

		for {
			innerApp := WorkflowApp{}
			_, err := it.Next(&innerApp)
			if err != nil {
				if strings.Contains(fmt.Sprintf("%s", err), "cannot load field") {
					log.Printf("[WARNING] Error in public app load: %s.", err)
					continue
				}

				log.Printf("[WARNING] No more apps (public)? Breaking: %s.", err)
				break
			}

			//log.Printf("APP: %s", innerApp.Name)
			found := false
			//log.Printf("ACTIONS: %d - %s", len(app.Actions), app.Name)
			for _, loopedApp := range allApps {
				if loopedApp.Name == innerApp.Name || loopedApp.ID == innerApp.ID {
					found = true
					break
				}
			}

			if !found {
				allApps = append(allApps, innerApp)
			}
		}

		if err != iterator.Done {
			//log.Printf("[INFO] Failed fetching results: %v", err)
			//break
		}

		// Get the cursor for the next page of results.
		nextCursor, err := it.Cursor()
		if err != nil {
			log.Printf("Cursorerror: %s", err)
			break
		} else {
			//log.Printf("NEXTCURSOR: %s", nextCursor)
			nextStr := fmt.Sprintf("%s", nextCursor)
			if cursorStr == nextStr {
				break
			}

			cursorStr = nextStr
			query = query.Start(nextCursor)
			//cursorStr = nextCursor
			//break
		}

		if len(allApps) > maxLen {
			break
		}
	}

	if len(allApps) > 0 {
		newbody, err := json.Marshal(allApps)
		if err != nil {
			return allApps, nil
		}

		err = SetCache(ctx, cacheKey, newbody)
		if err != nil {
			log.Printf("[INFO] Error setting app cache item for %s: %v", cacheKey, err)
		} else {
			log.Printf("[INFO] Set app cache for %s", cacheKey)
		}
	}

	return allApps, nil
}

func GetAllWorkflowApps(ctx context.Context, maxLen int) ([]WorkflowApp, error) {
	var allApps []WorkflowApp

	wrapper := []WorkflowApp{}
	cacheKey := fmt.Sprintf("workflowapps-sorted-%d", maxLen)
	if project.CacheDb {
		cache, err := GetCache(ctx, cacheKey)
		if err == nil {
			cacheData := []byte(cache.([]uint8))
			err = json.Unmarshal(cacheData, &wrapper)
			if err == nil {
				return wrapper, nil
			}
		} else {
			//log.Printf("[INFO] Failed getting cache for apps with KEY %s: %s", cacheKey, err)
		}
	}

	cursorStr := ""
	query := datastore.NewQuery("workflowapp").Order("-edited").Limit(10)
	//query := datastore.NewQuery("workflowapp").Order("-edited").Limit(40)

	// NOT BEING UPDATED
	// FIXME: Update the app with the correct actions. HOW DOES THIS WORK??
	// Seems like only actions are wrong. Could get the app individually.
	// Guessing it's a memory issue.
	//var err error
	for {
		it := project.Dbclient.Run(ctx, query)
		//innerApp := WorkflowApp{}
		//data, err := it.Next(&innerApp)
		//log.Printf("DATA: %#v, err: %s", data, err)

		for {
			innerApp := WorkflowApp{}
			_, err := it.Next(&innerApp)
			if err != nil {
				//log.Printf("No more apps? Breaking: %s.", err)
				break
			}

			if innerApp.Name == "Shuffle Subflow" {
				continue
			}

			if !innerApp.IsValid {
				continue
			}

			found := false
			//log.Printf("ACTIONS: %d - %s", len(app.Actions), app.Name)
			for _, loopedApp := range allApps {
				if loopedApp.Name == innerApp.Name {
					found = true
					break
				}
			}

			if !found {
				allApps = append(allApps, innerApp)
			}
		}

		if err != iterator.Done {
			//log.Printf("[INFO] Failed fetching results: %v", err)
			//break
		}

		// Get the cursor for the next page of results.
		nextCursor, err := it.Cursor()
		if err != nil {
			log.Printf("Cursorerror: %s", err)
			break
		} else {
			//log.Printf("NEXTCURSOR: %s", nextCursor)
			nextStr := fmt.Sprintf("%s", nextCursor)
			if cursorStr == nextStr {
				break
			}

			cursorStr = nextStr
			query = query.Start(nextCursor)
			//cursorStr = nextCursor
			//break
		}

		if len(allApps) > maxLen {
			break
		}
	}

	//log.Printf("FOUND %d apps", len(allApps))
	if project.CacheDb {
		log.Printf("[INFO] Setting %d apps in cache for 10 minutes for %s", len(allApps), cacheKey)

		//requestCache.Set(cacheKey, &apps, cache.DefaultExpiration)
		data, err := json.Marshal(allApps)
		if err == nil {
			err = SetCache(ctx, cacheKey, data)
			if err != nil {
				log.Printf("[WARNING] Failed updating cache for execution: %s", err)
			}
		} else {
			log.Printf("[WARNING] Failed marshalling execution: %s", err)
		}
	}

	//var allworkflowapps []WorkflowApp
	//_, err := dbclient.GetAll(ctx, query, &allworkflowapps)
	//if err != nil {
	//	if strings.Contains(fmt.Sprintf("%s", err), "ResourceExhausted") {
	//		//datastore.NewQuery("workflowapp").Limit(30).Order("-edited")
	//		query = datastore.NewQuery("workflowapp").Order("-edited").Limit(25)
	//		//q := q.Limit(25)
	//		_, err := dbclient.GetAll(ctx, query, &allworkflowapps)
	//		if err != nil {
	//			return []WorkflowApp{}, err
	//		}
	//	} else {
	//		return []WorkflowApp{}, err
	//	}
	//}

	return allApps, nil
}

func SetWorkflowQueue(ctx context.Context, executionRequests ExecutionRequestWrapper, id string) error {
	key := datastore.NameKey("workflowqueue", id, nil)

	// New struct, to not add body, author etc
	if _, err := project.Dbclient.Put(ctx, key, &executionRequests); err != nil {
		log.Printf("Error adding workflow queue: %s", err)
		return err
	}

	return nil
}

func GetWorkflowQueue(ctx context.Context, id string) (ExecutionRequestWrapper, error) {

	key := datastore.NameKey("workflowqueue", id, nil)
	workflows := ExecutionRequestWrapper{}
	if err := project.Dbclient.Get(ctx, key, &workflows); err != nil {
		return ExecutionRequestWrapper{}, err
	}

	return workflows, nil
}

func SetWorkflow(ctx context.Context, workflow Workflow, id string, optionalEditedSecondsOffset ...int) error {
	workflow.Edited = int64(time.Now().Unix())
	if len(optionalEditedSecondsOffset) > 0 {
		workflow.Edited += int64(optionalEditedSecondsOffset[0])
	}

	key := datastore.NameKey("workflow", id, nil)

	// New struct, to not add body, author etc
	if _, err := project.Dbclient.Put(ctx, key, &workflow); err != nil {
		log.Printf("Error adding workflow: %s", err)
		return err
	}

	return nil
}

func SetWorkflowAppAuthDatastore(ctx context.Context, workflowappauth AppAuthenticationStorage, id string) error {
	key := datastore.NameKey("workflowappauth", id, nil)

	// New struct, to not add body, author etc
	if _, err := project.Dbclient.Put(ctx, key, &workflowappauth); err != nil {
		log.Printf("[WARNING] Error adding workflow app AUTH: %s", err)
		return err
	}

	return nil
}

func SetEnvironment(ctx context.Context, data *Environment) error {
	// clear session_token and API_token for user
	k := datastore.NameKey("Environments", strings.ToLower(data.Name), nil)

	// New struct, to not add body, author etc

	if _, err := project.Dbclient.Put(ctx, k, data); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetSchedule(ctx context.Context, schedulename string) (*ScheduleOld, error) {
	key := datastore.NameKey("schedules", strings.ToLower(schedulename), nil)
	curUser := &ScheduleOld{}
	if err := project.Dbclient.Get(ctx, key, curUser); err != nil {
		return &ScheduleOld{}, err
	}

	return curUser, nil
}

func GetApikey(ctx context.Context, apikey string) (User, error) {
	// Query for the specific API-key in users
	q := datastore.NewQuery("Users").Filter("apikey =", apikey)
	var users []User
	_, err = project.Dbclient.GetAll(ctx, q, &users)
	if err != nil {
		log.Printf("[WARNING] Error getting apikey: %s", err)
		return User{}, err
	}

	if len(users) == 0 {
		log.Printf("[WARNING] No users found for apikey %s", apikey)
		return User{}, err
	}

	return users[0], nil
}

func GetHook(ctx context.Context, hookId string) (*Hook, error) {
	key := datastore.NameKey("hooks", strings.ToLower(hookId), nil)
	hook := &Hook{}
	if err := project.Dbclient.Get(ctx, key, hook); err != nil {
		return &Hook{}, err
	}

	return hook, nil
}

func SetHook(ctx context.Context, hook Hook) error {
	key1 := datastore.NameKey("hooks", strings.ToLower(hook.Id), nil)

	// New struct, to not add body, author etc
	if _, err := project.Dbclient.Put(ctx, key1, &hook); err != nil {
		log.Printf("Error adding hook: %s", err)
		return err
	}

	return nil
}

func GetFile(ctx context.Context, id string) (*File, error) {
	key := datastore.NameKey("Files", id, nil)
	curFile := &File{}
	if err := project.Dbclient.Get(ctx, key, curFile); err != nil {
		return &File{}, err
	}

	return curFile, nil
}

func SetFile(ctx context.Context, file File) error {
	// clear session_token and API_token for user
	timeNow := time.Now().Unix()
	file.UpdatedAt = timeNow

	k := datastore.NameKey("Files", file.Id, nil)
	if _, err := project.Dbclient.Put(ctx, k, &file); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetAllFiles(ctx context.Context, orgId string) ([]File, error) {
	var files []File
	q := datastore.NewQuery("Files").Filter("org_id =", orgId).Limit(100)

	_, err := project.Dbclient.GetAll(ctx, q, &files)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%s", err), "ResourceExhausted") {
			q = q.Limit(50)
			_, err := project.Dbclient.GetAll(ctx, q, &files)
			if err != nil {
				return []File{}, err
			}
		} else if strings.Contains(fmt.Sprintf("%s", err), "cannot load field") {
			log.Printf("[INFO] Failed loading SOME files - skipping: %s", err)
		} else {
			return []File{}, err
		}
	}

	return files, nil
}

func GetWorkflowAppAuthDatastore(ctx context.Context, id string) (*AppAuthenticationStorage, error) {

	key := datastore.NameKey("workflowappauth", id, nil)
	appAuth := &AppAuthenticationStorage{}
	// New struct, to not add body, author etc
	if err := project.Dbclient.Get(ctx, key, appAuth); err != nil {
		return &AppAuthenticationStorage{}, err
	}

	return appAuth, nil
}

func GetAllSchedules(ctx context.Context, orgId string) ([]ScheduleOld, error) {
	var schedules []ScheduleOld

	q := datastore.NewQuery("schedules").Filter("org = ", orgId)
	//CreatedAt    int64    `json:"created_at" datastore:"created_at"`
	if orgId == "ALL" {
		q = datastore.NewQuery("schedules")
	}

	_, err := project.Dbclient.GetAll(ctx, q, &schedules)
	if err != nil {
		return []ScheduleOld{}, err
	}

	return schedules, nil
}
