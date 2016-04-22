package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"strconv"
	"strings"
	"time"
)

// define static config constants
const MONGOIP string = "127.0.0.1"

// Top-Level Flag Definitions
var addhost = flag.Bool("addHost", false, "Use this flag to add a host to the provisioner database")
var deletehost = flag.Bool("deleteHost", false, "Use this flag to delete a host from the provisioner database")
var attachhost = flag.Bool("attachHost", false, "Use this flag to attach a host to a group")
var detachhost = flag.Bool("detachHost", false, "Use this flag to remove a host from a group")
var movehost = flag.Bool("moveHost", false, "Use this flag to move a host from one Ansible group to another")
var addgroup = flag.Bool("addGroup", false, "Use this flag to add a group to the provisioner database")
var deletegroup = flag.Bool("deleteGroup", false, "Use this flag to delete a group from the provisioner database")
var clonehost = flag.Bool("cloneHost", false, "Use this flag to clone a host")
var addenv = flag.Bool("addEnvironment", false, "Use this flag to delete a group from the provisioner database")
var push = flag.Bool("push", false, "Use this flag to push a host to the custodian database")
var pull = flag.Bool("pull", false, "Use this flag to pull a host from the custodian database")
var listgroups = flag.Bool("listGroups", false, "Use this flag to list the groups from the specified database")
var listhostopts = flag.Bool("hostOptions", false, "Use this flag to list the hosts from the specified database")
var listgroupopts = flag.Bool("groupOptions", false, "Use this flag to list the groups from the specified database")
var hostdetails = flag.Bool("hostDetails", false, "Use this flag to display the host details for a given host")
var groupdetails = flag.Bool("groupDetails", false, "Use this flag to display the group details for a given group")

// Sub-Flag Definitions
var fqdn = flag.String("fqdn", "EMPTY", "Fully Qualified Domain Name of the host")
var groups = flag.String("groups", "EMPTY", "The Name of an Ansible Group")
var hosts = flag.String("hosts", "EMPTY", "A comma delimited list of hostnames or a single hostname")
var group = flag.String("group", "EMPTY", "The Name of an Ansible Group")
var template = flag.String("template", "EMPTY", "The Name of the template to use to create clone")
var clone = flag.String("clone", "EMPTY", "The name of the clone to be added to the provisioner database")
var environment = flag.String("environment", "EMPTY", "The name of an Ansible Compute Environment")
var description = flag.String("description", "EMPTY", "A Short Description of the Ansible Group")
var togroup = flag.String("to-group", "EMPTY", "A valid Ansible group")
var fromgroup = flag.String("from-group", "EMPTY", "A valid Ansible group")
var brepoversion = flag.String("baseRepoVersion", "EMPTY", "A valid Pulp Base repository version")
var urepoversion = flag.String("updatesRepoVersion", "EMPTY", "A valid Pulp Updates repository version")
var erepoversion = flag.String("extrasRepoVersion", "EMPTY", "A valid Pulp Extras repository version")
var prepoversion = flag.String("plusRepoVersion", "EMPTY", "A valid Pulp Plus repository version")
var eplrepoversion = flag.String("epelRepoVersion", "EMPTY", "A valid Pulp Epel repository version")
var ostype = flag.String("osType", "EMPTY", "Operating System Type (e.g, CentOS|RedHat)")
var osversion = flag.String("osVersion", "EMPTY", "Operating System Version (e.g, 7.0)")
var machinearch = flag.String("archType", "EMPTY", "Machine Architecture Type (e.g, x86_64)")
var datastore = flag.String("datastore", "EMPTY", "Datastore name to run against (e.g, provisioner)")
var region = flag.String("region", "EMPTY", "Geographical region to run against (Atlanta)")

// Type Definitions
type AnsibleGroups struct {
	Members     map[string][]string
	Description string
	Environment string
	Name        string
}

type AnsibleHostMeta struct {
	VarMap map[string]string
}

type PulpClient struct {
	Fqdn        string
	RpmRepos    map[string]string
	OsType      string
	OsVersion   string
	MachineArch string
}

type AnsibleHost struct {
	Fqdn        string
	Groups      map[string]bool
	Environment string
}

type AnsibleEnvironment struct {
	Name   string
	Prefix string
	Groups map[string]bool
}

type AnsibleHostVars struct {
	Name   string
	VarMap map[string]string
}

type InventoryFile struct {
	Path        string
	Environment string
}

func main() {

	if len(os.Args) < 2 {
		listInventory()
		os.Exit(0)

	}

	// Ensure that argument list constains the required parameter or do nothing
	if os.Args[1] == "--list" {
		listInventory()
		os.Exit(0)
	}

	if os.Args[1] == "--host" {
		if len(os.Args) != 3 {
			fmt.Println("\n[ ERROR ] --> This switch requires a single parameter that is a hostname.\n")
			os.Exit(1)
		}

		listHostVars()
		os.Exit(0)
	}

	if os.Args[1] == "--add-host" {
		// get Database from stdin
		dbReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the database to which you wish to add the host: ")
		dbName, _ := dbReader.ReadString('\n')
		dBase := strings.Trim(dbName, "\n")

		// get host FQDN from stdin
		fqdnReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter FQDN of host you wish to add: ")
		fqdName, _ := fqdnReader.ReadString('\n')
		fName := strings.Trim(fqdName, "\n")

		// get host OS Type from stdin
		osTypeReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter The Type of Operating System of the host you wish to add (RedHat|CentOS): ")
		oType, _ := osTypeReader.ReadString('\n')
		osType := strings.Trim(oType, "\n")

		// get host OS Version from stdin
		osVersionReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter The Version of Operating System of the host you wish to add (e.g, 7.0): ")
		oVersion, _ := osVersionReader.ReadString('\n')
		osVersion := strings.Trim(oVersion, "\n")

		// get host Machine Architecture from stdin
		machineArchReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter The  Machine Architecture of the host you wish to add (e.g, x86_64): ")
		mArch, _ := machineArchReader.ReadString('\n')
		machArch := strings.Trim(mArch, "\n")

		// get Pulp Base  Repository version of host from stdin
		baseRepoReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter version of the Base repository (Enter 1 if uncertain) : ")
		bRepo, _ := baseRepoReader.ReadString('\n')
		baseRepo := strings.Trim(bRepo, "\n")

		// get Pulp Update  Repository version of host from stdin
		updateRepoReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter version of the Update repository (Enter 1 if uncertain) : ")
		uRepo, _ := updateRepoReader.ReadString('\n')
		updatesRepo := strings.Trim(uRepo, "\n")

		// get Pulp Extras  Repository version of host from stdin
		extrasRepoReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter version of the Extras Repository (Enter 1 if uncertain) : ")
		eRepo, _ := extrasRepoReader.ReadString('\n')
		extrasRepo := strings.Trim(eRepo, "\n")

		// get Pulp Plus  Repository version of host from stdin
		plusRepoReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter version of the Plus Repository (Enter 1 if uncertain) : ")
		pRepo, _ := plusRepoReader.ReadString('\n')
		plusRepo := strings.Trim(pRepo, "\n")

		// get Pulp Epel  Repository version of host from stdin
		epelRepoReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter version of the Epel Repository (Enter 1 if uncertain) : ")
		eplRepo, _ := epelRepoReader.ReadString('\n')
		epelRepo := strings.Trim(eplRepo, "\n")

		// validate environment
		if !envExists(ENV, dBase) {
			fmt.Println("\n[ FAILED ] --> The environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		// validate host
		if hostExists(fName, ENV, dBase) {
			fmt.Println("\n[ FAILED ] --> The Host: " + fName + " already exists in Environment: " + ENV + ".\n")
			os.Exit(1)
		}

		// ensure that the data store has been provided
		if dBase != "provisioner" && dBase != "custodian" {
			fmt.Println("\n[ ERROR ] --> You did not provide a valid value for the datastore when it is required.\n")
			os.Exit(1)
		}

		groupsMap := make(map[string]bool)

		aHost := AnsibleHost{Fqdn: fName, Groups: groupsMap, Environment: ENV}
		// adding host to datastore -- should never have a host added to both datastores at the same time.
		addHost(aHost, dBase)

		repoMap := make(map[string]string)
		repoMap["Base"] = baseRepo
		repoMap["Updates"] = updatesRepo
		repoMap["Extras"] = extrasRepo
		repoMap["Plus"] = plusRepo
		repoMap["Epel"] = epelRepo

		pHost := PulpClient{Fqdn: fName, RpmRepos: repoMap, OsType: osType, OsVersion: osVersion, MachineArch: machArch}
		addPulpClient(pHost, dBase)
		fmt.Println("\n[ OK ] --> Successfully added " + pHost.Fqdn + " to the Pulp database.\n")

		fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in " + dBase + "...............\n")
		updateInventoryFile(ENV, dBase)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + "\n")

		os.Exit(0)
	}

	if os.Args[1] == "--list-groups" {

		// validate Environment
		if !envExists(ENV, "provisioner") {
			fmt.Println("\n[ FAILED ] --> The environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		// list groups and their description that exist in the supplied environment
		listGroups(ENV, "provisioner")
		os.Exit(0)
	}

	if os.Args[1] == "--display-host" {

		// ensure that the data store has been provided
		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		// get hostname (fqdn) from stdin
		fqdnReader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter FQDN of host: ")
		hostName, _ := fqdnReader.ReadString('\n')
		hName := strings.Trim(hostName, "\n")

		// validate host
		if !hostExists(hName, ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The Host: " + hName + " does not exist in Environment: " + ENV + ".\n")
			os.Exit(1)
		}

		displayHost(hName, ENV, *datastore)
		os.Exit(0)
	}

	if os.Args[1] == "--group-options" {
		// Get list of groups
		listGroupOptions(ENV, "provisioner")
		os.Exit(0)
	}

	if os.Args[1] == "--host-options" {
		//Get list of hosts
		listHostOptions(ENV, *datastore)
		os.Exit(0)
	}

	if os.Args[1] == "--add-group" {
		// get Database from stdin
		dbReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the database to which you wish to add the host: ")
		dbName, _ := dbReader.ReadString('\n')
		dBase := strings.Trim(dbName, "\n")

		// get group name from stdin
		groupNameReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter Name of group to add : ")
		groupName, _ := groupNameReader.ReadString('\n')
		gName := strings.Trim(groupName, "\n")

		// get group description from stdin
		groupReader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter short group description: ")
		groupDescription, _ := groupReader.ReadString('\n')
		gDesc := strings.Trim(groupDescription, "\n")

		// setup the group members map with empty members slice
		groupMembers := map[string][]string{gName: make([]string, 0)}
		aGroup := AnsibleGroups{Members: groupMembers, Description: gDesc, Environment: ENV, Name: gName}

		if dBase == "all" {
			if !envExists(ENV, "provisioner") || !envExists(ENV, "custodian") {
				fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in all databases.\n")
				fmt.Println("\n[ FATAL ERROR ] --> Please Ensure that your environments are set up correctly....Exiting.\n")
				os.Exit(1)

			}
			if groupExists(gName, ENV, "provisioner") && groupExists(gName, ENV, "custodian") {
				fmt.Println("\n[ ERROR ] --> Group: " + gName + " already exists in all databases...skipping add.\n")
				os.Exit(1)
			}

			if groupExists(gName, ENV, "provisioner") {
				fmt.Println("\n[ INFO ] --> group: " + gName + " already exists in provisioner...skipping add.\n")
			} else {
				// Add the group to the requested environment in provisioner datastore
				addGroup(aGroup, "provisioner")
				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in provisioner...............\n")
				// Update Inventory File
				updateInventoryFile(ENV, "provisioner")
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in provisioner\n")
			}

			if groupExists(gName, ENV, "custodian") {
				fmt.Println("\n[ INFO ] --> group: " + gName + " already exists in custodian...skipping add.\n")
			} else {
				// Add the group to the requested environment in custodian datastore
				addGroup(aGroup, "custodian")
				fmt.Println("\n[ INFO] --> Updating Inventory file for Environment: " + ENV + " in custodian...............\n")
				// Update Inventory File
				updateInventoryFile(ENV, "custodian")
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in custodian.\n")
			}

			os.Exit(0)

		} else {
			if !envExists(ENV, dBase) {
				fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database: " + dBase + ".\n")
				os.Exit(1)
			}

			if groupExists(gName, ENV, dBase) {
				fmt.Println("\n[ WARNING ] --> The group: " + gName + " already exists in the datastore: " + dBase + "...skipping add.\n")
				os.Exit(0)
			} else {
				// Add the group the requested environment in the specified datastore
				addGroup(aGroup, dBase)
				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in database: " + dBase + "...............\n")
				// Update Inventory File
				updateInventoryFile(ENV, dBase)
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in database: " + dBase + ".\n")

				os.Exit(0)

			}
		}
	}

	if os.Args[1] == "--attach-host" {
		// get Database from stdin
		dbReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the database to which you wish to add the host: ")
		dbName, _ := dbReader.ReadString('\n')
		dBase := strings.Trim(dbName, "\n")

		// get group from stdin
		hostReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the FQDN of the host you wish to attach: ")
		hostName, _ := hostReader.ReadString('\n')
		hName := strings.Trim(hostName, "\n")

		// get group from stdin
		groupReader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the group to wich you wish to attach this host: ")
		groupName, _ := groupReader.ReadString('\n')
		gName := strings.Trim(groupName, "\n")

		// ensure that the data store has been provided
		if dBase != "provisioner" && dBase != "custodian" {
			fmt.Println("\n[ ERROR ] --> You did not provide a valid value for the datastore when it is required.\n")
			os.Exit(1)
		}

		// validate environment
		if !envExists(ENV, dBase) {
			fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		// validate host
		if !hostExists(hName, ENV, dBase) {
			fmt.Println("\n[ ERROR ] --> The Host: " + hName + " does not exist in Environment: " + ENV + " in database: " + dBase + ".\n")
			fmt.Println("There is nothing to delete...Exiting.\n")
			os.Exit(1)
		}

		// validate group
		if !groupExists(gName, ENV, dBase) {
			fmt.Println("\n[ ERROR ] --> The Group: " + gName + " does not exist in Environment: " + ENV + " in database: " + dBase + ".\n")
			os.Exit(1)
		}

		// attach supplied host to the requested group
		attachHost(hName, gName, ENV, dBase)

		// Update Inventory File
		fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in database: " + dBase + "...............\n")
		updateInventoryFile(ENV, dBase)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in database: " + dBase + ".\n")

		os.Exit(0)
	}

	if os.Args[1] == "--detach-host" {
		// get Database from stdin
		dbReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the database to which you wish to add the host: ")
		dbName, _ := dbReader.ReadString('\n')
		dBase := strings.Trim(dbName, "\n")

		// get group from stdin
		hostReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the FQDN of the host you wish to detach: ")
		hostName, _ := hostReader.ReadString('\n')
		hName := strings.Trim(hostName, "\n")

		// get group from stdin
		groupReader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the group from which you wish to detach this host: ")
		groupName, _ := groupReader.ReadString('\n')
		gName := strings.Trim(groupName, "\n")

		// ensure that the data store has been provided
		if dBase != "provisioner" && dBase != "custodian" && dBase != "all" {
			fmt.Println("\n[ ERROR ] --> You did not provide a valid value for the datastore when it is required.\n")
			os.Exit(1)
		}

		// validate environment
		if !envExists(ENV, dBase) {
			fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		if dBase == "all" {
			// validate group
			if !groupExists(gName, ENV, "provisioner") {
				fmt.Println("\n[ FAILED ] --> The Group: " + gName + " does not exist in Environment: " + ENV + " in datastore: provisioner.\n")
				os.Exit(1)
			}

			if !groupExists(gName, ENV, "custodian") {
				fmt.Println("\n[ FAILED ] --> The Group: " + gName + " does not exist in Environment: " + ENV + " in datastore: custodian.\n")
				os.Exit(1)
			}

			// validate host
			if !hostExists(hName, ENV, "provisioner") {
				fmt.Println("\n[ FAILED ] --> The Host: " + hName + " does not exist in Environment: " + ENV + " in provisioner.\n")
				os.Exit(1)
			}

			// validate host
			if !hostExists(hName, ENV, "custodian") {
				fmt.Println("\n[ FAILED ] --> The Host: " + hName + " does not exist in Environment: " + ENV + " in custodian.\n")
				os.Exit(1)
			}

			// detach supplied host from the supplied group in all datastores
			fmt.Println("\nDetaching host: " + hName + " from group: " + gName + " in provisioner............\n")
			detachGroupFromHost(hName, gName, ENV, "provisioner")
			detachHostFromGroup(hName, gName, ENV, "provisioner")
			fmt.Println("\n[ OK] --> Successfully detached host: " + hName + " from group: " + gName + " in provisioner...........\n")
			fmt.Println("\nDetaching host: " + hName + " from group: " + gName + " in custodian............\n")
			detachGroupFromHost(hName, gName, ENV, "custodian")
			detachHostFromGroup(hName, gName, ENV, "custodian")
			fmt.Println("\n[ OK] --> Successfully detached host: " + hName + " from group: " + gName + " in custodian............\n")

			fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " for provisioner...............\n")
			// Update Inventory File
			updateInventoryFile(ENV, "provisioner")
			fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in provisioner\n")

			fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " for custodian...............\n")
			// Update Inventory File
			updateInventoryFile(ENV, "custodian")
			fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in custodian\n")

			os.Exit(0)

		} else {
			if !groupExists(gName, ENV, dBase) {
				fmt.Println("\n[ FAILED ] --> The Group: " + gName + " does not exist in Environment: " + ENV + " in datastore: " + dBase + ".\n")
				os.Exit(1)
			}

			// validate host
			if !hostExists(hName, ENV, dBase) {
				fmt.Println("\n[ FAILED ] --> The Host: " + hName + " does not exist in Environment: " + ENV + " in " + dBase + ".\n")
				os.Exit(1)
			}

			// detach supplied host from the supplied group in the datastore
			fmt.Println("\nDetaching host: " + hName + " from group: " + gName + " from datastore: " + dBase + "............\n")
			detachGroupFromHost(hName, gName, ENV, dBase)
			detachHostFromGroup(hName, gName, ENV, dBase)
			fmt.Println("\n[ OK] --> Successfully detached host: " + hName + " from group: " + gName + " in datastore: " + dBase + "............\n")

			fmt.Println("\nUpdating Inventory file for Environment: " + ENV + "...............\n")
			// Update Inventory File
			updateInventoryFile(ENV, dBase)
			fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + "\n")

			os.Exit(0)

		}
	}

	if os.Args[1] == "--delete-host" {

		// get group from stdin
		hostReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the FQDN of the host you wish to delete: ")
		hostName, _ := hostReader.ReadString('\n')
		hName := strings.Trim(hostName, "\n")

		// get Database from stdin
		dbReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the database to which you wish to add the host: ")
		dbName, _ := dbReader.ReadString('\n')
		dBase := strings.Trim(dbName, "\n")

		// ensure that the data store has been provided
		if dBase != "provisioner" && dBase != "custodian" && dBase != "all" {
			fmt.Println("\n[ ERROR ] --> You did not provide a valid value for the datastore when it is required.\n")
			os.Exit(1)
		}

		if !envExists(ENV, dBase) {
			fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		if !hostExists(hName, ENV, dBase) {
			fmt.Println("\n[ FAILED ] --> The Host: " + hName + " does not exist in Environment: " + ENV + ".\n")
			os.Exit(1)
		}

		// delete supplied host from the supplied group
		fmt.Println("\nDeleting host: " + hName + "............\n")
		deleteHost(hName, ENV, dBase)
		fmt.Println("\n[ OK ] --> Successfully deleted host: " + hName + "\n")

		// delete supplied host from the Pulp repository database
		fmt.Println("\nDeleting host: " + hName + " from the Pulp Repository database............\n")
		deletePulpClient(hName, dBase)

		fmt.Println("\n[ INFO] --> Updating Inventory file for Environment: " + ENV + "...............\n")
		// Update Inventory File
		updateInventoryFile(ENV, dBase)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + "\n")

		os.Exit(0)
	}

	if os.Args[1] == "--delete-group" {

		// get group from stdin
		groupReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the Groupname of the Group you wish to delete: ")
		groupName, _ := groupReader.ReadString('\n')
		gName := strings.Trim(groupName, "\n")

		// get Database from stdin
		dbReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the database to which you wish to add the host: ")
		dbName, _ := dbReader.ReadString('\n')
		dBase := strings.Trim(dbName, "\n")

		// ensure that the data store has been provided
		if dBase != "provisioner" && dBase != "custodian" && dBase != "all" {
			fmt.Println("\n[ ERROR ] --> You did not provide a valid value for the datastore when it is required.\n")
			os.Exit(1)
		}

		if dBase == "all" {
			// validate environment
			if !envExists(ENV, "privisioner") || !envExists(ENV, "custodian") {
				fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in all databases.\n")
				fmt.Println("\n[ FATAL ERROR ] --> Please Ensure that your environments are set up correctly....Exiting.\n")
				os.Exit(1)
			}

			// delete supplied group from the supplied environment
			if groupExists(gName, ENV, "provisioner") {
				fmt.Println("\n[ INFO ] --> Deleting group: " + gName + "from Environment: " + ENV + " in provisioner............\n")
				deleteGroup(gName, ENV, "provisioner")
				fmt.Println("\n[ OK ] --> Successfully deleted group: " + gName + " from Environment: " + ENV + " in provisioner.\n")

				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in provisioner...............\n")
				// Update Inventory File
				updateInventoryFile(ENV, "provisioner")
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in provisioner.\n")
			} else {
				fmt.Println("\n[ INFO ] --> Group: " + gName + " does not exist in datastore: provisioner...skipping delete.\n")
			}

			if groupExists(gName, ENV, "custodian") {
				fmt.Println("\n[ INFO ] --> Deleting group: " + gName + "from Environment: " + ENV + " in custodian............\n")
				deleteGroup(gName, ENV, "custodian")
				fmt.Println("\n[ OK ] --> Successfully deleted group: " + gName + " from Environment: " + ENV + "in custodian.\n")

				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in custodian...............\n")
				// Update Inventory File
				updateInventoryFile(ENV, "custodian")
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in custodian.\n")

			} else {
				fmt.Println("\n[ INFO ] --> Group: " + gName + " does not exist in datastore: custodian...skipping delete.\n")
			}

			os.Exit(0)

		} else {
			// validate environment
			if !envExists(ENV, dBase) {
				fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database: " + dBase + ".\n")
				os.Exit(1)
			}
			// delete supplied group from the supplied environment
			if groupExists(gName, ENV, dBase) {
				fmt.Println("\n[ INFO ] --> Deleting group: " + gName + "from Environment: " + ENV + " in datastore: " + dBase + "............\n")
				deleteGroup(gName, ENV, dBase)
				fmt.Println("\n[ OK ] --> Successfully deleted group: " + gName + " from Environment: " + ENV + " in datastore: " + dBase + ".\n")

				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in " + dBase + "...............\n")
				// Update Inventory File
				updateInventoryFile(ENV, dBase)
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + dBase + ".\n")

				os.Exit(0)
			} else {
				fmt.Println("\n[ ERROR ] --> Failed to delete group: " + gName + " from " + dBase + "...group not found.\n")
				os.Exit(1)
			}

		}

	}

	if os.Args[1] == "--clone-host" {

		// get group from stdin
		templateReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the FQDN of the host to use as a template: ")
		templateName, _ := templateReader.ReadString('\n')
		tName := strings.Trim(templateName, "\n")

		// get group from stdin
		hostReader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the FQDN of the host you wish to add: ")
		hostName, _ := hostReader.ReadString('\n')
		hName := strings.Trim(hostName, "\n")

		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		// create the new host using the supplied template host
		cloneHost(tName, hName, ENV, *datastore)

		fmt.Println("\nUpdating Inventory file for Environment: " + ENV + " in database: " + *datastore + "...............\n")
		// Update Inventory File
		updateInventoryFile(ENV, *datastore)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in database: " + *datastore + ".\n")

		os.Exit(0)
	}

	if os.Args[1] == "--add-env" {
		// get database from stdin
		dbReader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter the name of the database you wish to add the environment to: ")
		dbName, _ := dbReader.ReadString('\n')
		dBase := strings.Trim(dbName, "\n")

		if envExists(ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The environment: " + ENV + " already Exists in the database.\n")
			os.Exit(1)
		}

		// create prefix
		ePrefix := strings.ToLower(strings.Replace(ENV, "-", "_", -1))

		// create AnsibleEnvironment struct and populate fields
		anEnvironment := new(AnsibleEnvironment)
		anEnvironment.Prefix = ePrefix
		anEnvironment.Name = ENV

		// attach supplied host to the requested group
		addEnvironment(anEnvironment, dBase)

		fmt.Println("\n[ INFO ] --> Creating Inventory file for Environment: " + ENV + "...............\n")
		// add Inventory file
		createInventoryFile(anEnvironment.Name, dBase)
		fmt.Println("\n[ OK ] --> Successfully Created Inventory File for " + ENV + ".\n")

		os.Exit(0)
	}

	// Parse Flags -- Should only get here if not an interactive run
	flag.Parse()

	// Check what flags were supplied to determine function and ensure proper subflags were also supplied
	if *addhost {
		//ensure necessary sub-flag values were supplied
		if *fqdn == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -fqdn when it is required.\n")
			os.Exit(1)
		}

		if *brepoversion == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -baseRepoVersion when it is required.\n")
			os.Exit(1)
		}

		if *urepoversion == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -updatesRepoVersion when it is required.\n")
			os.Exit(1)
		}

		if *erepoversion == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -extrasRepoVersion when it is required.\n")
			os.Exit(1)
		}

		if *prepoversion == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -plusRepoVersion when it is required.\n")
			os.Exit(1)
		}

		if *eplrepoversion == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -epelRepoVersion when it is required.\n")
			os.Exit(1)
		}

		if *machinearch == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -archType when it is required.\n")
			os.Exit(1)
		}

		if *ostype == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -osType when it is required.\n")
			os.Exit(1)
		}

		if *osversion == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -osVersion when it is required.\n")
			os.Exit(1)
		}

		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		// validate environment
		if !envExists(ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The environment: " + ENV + " does not exist in the database: " + *datastore + ".\n")
			os.Exit(1)
		}

		// validate host
		if hostExists(*fqdn, ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The Host: " + *fqdn + " already exists in Environment: " + ENV + ".\n")
			os.Exit(1)
		}

		groupsMap := make(map[string]bool)

		aHost := AnsibleHost{Fqdn: *fqdn, Groups: groupsMap, Environment: ENV}
		// we add the host before checking groups
		addHost(aHost, *datastore)

		repoMap := make(map[string]string)
		repoMap["Base"] = *brepoversion
		repoMap["Updates"] = *urepoversion
		repoMap["Extras"] = *erepoversion
		repoMap["Plus"] = *prepoversion
		repoMap["Epel"] = *eplrepoversion

		pHost := PulpClient{Fqdn: *fqdn, RpmRepos: repoMap, OsType: *ostype, OsVersion: *osversion, MachineArch: *machinearch}
		// we add the pulp client before checking group validity as well
		addPulpClient(pHost, *datastore)
		fmt.Println("\n[ OK ] --> Successfully added " + pHost.Fqdn + " to the Pulp database.\n")

		if *groups != "EMPTY" {
			if strings.Contains(*groups, ",") {
				gList := strings.Split(*groups, ",")
				for g := range gList {
					// ensure group exists in the specified environment in the specified datastore before proceeding
					if !groupExists(gList[g], ENV, *datastore) {
						fmt.Println("\n[ ERROR ] --> The Group: " + gList[g] + " does not exist in Environment: " + ENV + " in datastore: " + *datastore + ".\n")
						fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
						updateInventoryFile(ENV, *datastore)
						fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")
						os.Exit(1)
					}

					attachHost(*fqdn, gList[g], ENV, *datastore)
					fmt.Println("\n[ INFO ] -->Updating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
					updateInventoryFile(ENV, *datastore)
					fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")

				}
			} else if *groups != "" {
				if !groupExists(*groups, ENV, *datastore) {
					fmt.Println("\n[ ERROR ] --> The Group: " + *groups + " does not exist in Environment: " + ENV + " in datastore: " + *datastore + ".\n")
					os.Exit(1)
				}

				attachHost(*fqdn, *groups, ENV, *datastore)
			}
		}

		fmt.Println("\nUpdating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
		updateInventoryFile(ENV, *datastore)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")

		os.Exit(0)

	}

	if *listgroups {
		// validate environment
		if !envExists(ENV, *datastore) {
			fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		// validate Environment
		if !envExists(ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		// list groups and their description that exist in the supplied environment
		listGroups(ENV, *datastore)
		os.Exit(0)

	}

	if *listhostopts {
		// ensure that the data store has been provided
		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}
		//Get list of hosts
		listHostOptions(ENV, *datastore)
		os.Exit(0)
	}

	if *listgroupopts {
		// ensure that the data store has been provided
		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}
		// Get list of groups
		listGroupOptions(ENV, *datastore)
		os.Exit(0)
	}

        if *hostdetails {
                // ensure that the data store has been provided
                if *datastore == "EMPTY" {
                        fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
                        os.Exit(1)
                }

                // ensure that a fully qualified hostname has been provided
                if *fqdn == "EMPTY" {
                        fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -fqdn when it is required.\n")
                        os.Exit(1)
                }

                //display the host details
                displayHost(*fqdn, ENV, *datastore)
                os.Exit(0)
        }


        if *groupdetails {
                // ensure that the data store has been provided
                if *datastore == "EMPTY" {
                        fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
                        os.Exit(1)
                }

                // ensure that a group has been provided
                if *group == "EMPTY" {
                        fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -group when it is required.\n")
                        os.Exit(1)
                }

		 // validate group
                if !groupExists(*group, ENV, *datastore) {
                        fmt.Println("\n[ ERROR ] --> The Group: " + *group + " does not exist in Environment: " + ENV + " in datastore: " + *datastore + ".\n")
                        os.Exit(1)
                }


                //display the group details
                displayGroup(*group, ENV, *datastore)
                os.Exit(0)
        }


	if *attachhost {
		//ensure necessary sub-flag values were supplied
		if *fqdn == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -fqdn when it is required.\n")
			os.Exit(1)
		}

		if *groups == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -groups when it is required.\n")
			os.Exit(1)
		}

		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		// validate environment
		if !envExists(ENV, *datastore) {
			fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		// validate host
		if !hostExists(*fqdn, ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The Host: " + *fqdn + " does not exist in Environment: " + ENV + " in database: " + *datastore + ".\n")
			fmt.Println("There is nothing to delete...Exiting.\n")
			os.Exit(1)
		}

		if *groups != "EMPTY" {

			if strings.Contains(*groups, ",") {
				gList := strings.Split(*groups, ",")
				for g := range gList {
					if !groupExists(gList[g], ENV, *datastore) {
						fmt.Println("\n[ FAILED ] --> The Group: " + gList[g] + " does not exist in Environment: " + ENV + ".\n")

						// Update Inventory File
						fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
						updateInventoryFile(ENV, *datastore)
						fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")
						os.Exit(1)
					}

					attachHost(*fqdn, gList[g], ENV, *datastore)
				}
			} else if *groups != "" {
				if !groupExists(*groups, ENV, *datastore) {
					fmt.Println("\n[ FAILED ] --> The Group: " + *groups + " does not exist in Environment: " + ENV + " in " + *datastore + ".\n")
					os.Exit(1)
				}

				attachHost(*fqdn, *groups, ENV, *datastore)
			}
		} else {
			fmt.Println("\n[ FAILED ] -- > No group was supplied. You must supply a group to which to attach the host.\n")
			os.Exit(1)

		}

		// Update Inventory File
		fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
		updateInventoryFile(ENV, *datastore)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")

		os.Exit(0)
	}

	if *clonehost {

		//ensure necessary sub-flag values were supplied
		if *clone == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -clone when it is required.\n")
			os.Exit(1)
		}

		if *template == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -template when it is required.\n")
			os.Exit(1)
		}

		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		// create the new host using the supplied template host
		cloneHost(*template, *clone, ENV, *datastore)

		fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
		// Update Inventory File
		updateInventoryFile(ENV, *datastore)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")

		os.Exit(0)
	}

	if *deletehost {

		//ensure necessary sub-flag values were supplied
		if *fqdn == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -fqdn when it is required.\n")
			os.Exit(1)
		}

		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		// Ensure environment is valid
		if !envExists(ENV, *datastore) {
			fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		if !hostExists(*fqdn, ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The Host: " + *fqdn + " does not exist in Environment: " + ENV + ".\n")
			os.Exit(1)
		}

		// delete supplied host from the supplied group
		fmt.Println("\nDeleting host: " + *fqdn + "............\n")
		deleteHost(*fqdn, ENV, *datastore)
		fmt.Println("\n[ OK ] --> Successfully deleted host: " + *fqdn + "\n")

		// delete supplied host from the Pulp repository database
		fmt.Println("\nDeleting host: " + *fqdn + " from the Pulp Repository database............\n")
		deletePulpClient(*fqdn, *datastore)

		fmt.Println("\nUpdating Inventory file for Environment: " + ENV + " in database: " + *datastore + "...............\n")
		// Update Inventory File
		updateInventoryFile(ENV, *datastore)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in database: " + *datastore + ".\n")

		os.Exit(0)

	}

	if *detachhost {
		//ensure necessary sub-flag values were supplied
		if *fqdn == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -fqdn when it is required.\n")
			os.Exit(1)
		}

		if *groups == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -groups when it is required.\n")
			os.Exit(1)
		}

		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		// validate environment
		if !envExists(ENV, *datastore) {
			fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		// validate host
		if !hostExists(*fqdn, ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The Host: " + *fqdn + " does not exist in Environment: " + ENV + ".\n")
			os.Exit(1)
		}

		if *groups != "EMPTY" {

			if strings.Contains(*groups, ",") {
				gList := strings.Split(*groups, ",")
				for g := range gList {
					if !groupExists(gList[g], ENV, *datastore) {
						fmt.Println("\n[ FAILED ] --> The Group: " + gList[g] + " does not exist in Environment: " + ENV + ".\n")
						os.Exit(1)
					}

					detachGroupFromHost(*fqdn, gList[g], ENV, *datastore)
					detachHostFromGroup(*fqdn, gList[g], ENV, *datastore)
					fmt.Println("\n[ OK] --> Successfully detached host: " + *fqdn + " from group: " + gList[g] + " in " + *datastore + "............\n")
				}
			} else if *groups != "" {
				if !groupExists(*groups, ENV, *datastore) {
					fmt.Println("\n[ FAILED ] --> The Group: " + *groups + " does not exist in Environment: " + ENV + ".\n")
					os.Exit(1)
				}

				detachGroupFromHost(*fqdn, *groups, ENV, *datastore)
				detachHostFromGroup(*fqdn, *groups, ENV, *datastore)
				fmt.Println("\n[ OK] --> Successfully detached host: " + *fqdn + " from group: " + *groups + "............\n")
			}
		} else {
			fmt.Println("\n[ FAILED ] -- > No group was supplied. You must supply a group to which to attach the host.\n")
			os.Exit(1)

		}

		fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
		// Update Inventory File
		updateInventoryFile(ENV, *datastore)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")
		os.Exit(0)

	}

	if *push {
		//ensure necessary sub-flag values were supplied
		if *hosts == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -hosts when it is required.\n")
			os.Exit(1)
		}

		if *hosts != "EMPTY" {

			if strings.Contains(*hosts, ",") {
				hList := strings.Split(*hosts, ",")
				for h := range hList {
					if !hostExists(hList[h], ENV, "provisioner") {
						fmt.Println("\n[ FAILED ] --> The Host: " + hList[h] + " does not exist in Environment: " + ENV + ".\n")
						os.Exit(1)
					}

					pushOneHost(hList[h])
				}
			} else if *hosts != "" {
				if !hostExists(*hosts, ENV, "provisioner") {
					fmt.Println("\n[ FAILED ] --> The Host: " + *hosts + " does not exist in Environment: " + ENV + ".\n")
					os.Exit(1)
				}

				pushOneHost(*hosts)
			}
		}

		os.Exit(0)

	}

	if *pull {
		//ensure necessary sub-flag values were supplied
		if *hosts == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -hosts when it is required.\n")
			os.Exit(1)
		}

		if *hosts != "EMPTY" {

			if strings.Contains(*hosts, ",") {
				hList := strings.Split(*hosts, ",")
				for h := range hList {
					pullOneHost(hList[h])
				}
			} else if *hosts != "" {
				pullOneHost(*hosts)
			}
		}

		os.Exit(0)
	}

	if *addgroup {
		//ensure necessary sub-flag values were supplied
		if *group == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -group when it is required.\n")
			os.Exit(1)
		}

		if *description == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -description when it is required.\n")
			os.Exit(1)
		}

		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		// setup the group members map with empty members slice
		groupMembers := map[string][]string{*group: make([]string, 0)}
		aGroup := AnsibleGroups{Members: groupMembers, Description: *description, Environment: ENV, Name: *group}

		if *datastore == "all" {
			// validate environment
			if !envExists(ENV, "provisioner") || !envExists(ENV, "custodian") {
				fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in all databases.\n")
				fmt.Println("\n[ FATAL ERROR ] --> Please Ensure that your environments are set up correctly....Exiting.\n")
				os.Exit(1)
			}

			// validate group
			if groupExists(*group, ENV, "provisioner") && groupExists(*group, ENV, "custodian") {
				fmt.Println("\n[ FAILED ] --> The Group: " + *group + "  already exists in Environment: " + ENV + " in all datastores.\n")
				os.Exit(1)
			}
			if !groupExists(*group, ENV, "provisioner") {
				addGroup(aGroup, "provisioner")
				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in provisioner...............\n")
				// Update Inventory File
				updateInventoryFile(ENV, "provisioner")
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in provisioner.\n")

			}
			if !groupExists(*group, ENV, "custodian") {
				addGroup(aGroup, "custodian")
				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in custodian...............\n")
				// Update Inventory File
				updateInventoryFile(ENV, "custodian")
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in custodian.\n")

			}

			os.Exit(0)
		} else {

			if !envExists(ENV, *datastore) {
				fmt.Println("\n[ ERROR ] --> The environment: " + ENV + " does not exist in the datastore: " + *datastore + ".\n")
				os.Exit(1)
			}

			if !groupExists(*group, ENV, *datastore) {
				// Add the group the requested environment
				addGroup(aGroup, *datastore)
				fmt.Println("\nUpdating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
				// Update Inventory File
				updateInventoryFile(ENV, *datastore)
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")
				os.Exit(0)

			} else {
				fmt.Println("\n[ FAILED ] --> The Group: " + *group + "  already exists in Environment: " + ENV + " in database: " + *datastore + ".\n")
				os.Exit(1)
			}
		}
	}

	if *deletegroup {
		//ensure necessary sub-flag values were supplied
		if *group == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -group when it is required.\n")
			os.Exit(1)
		}

		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		if *datastore == "all" {
			// validate environment
			if !envExists(ENV, "provisioner") || !envExists(ENV, "custodian") {
				fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in all databases.\n")
				fmt.Println("\n[ FATAL ERROR ] --> Please Ensure that your environments are set up correctly....Exiting.\n")
				os.Exit(1)
			}
			if groupExists(*group, ENV, "provisioner") {
				// delete supplied group from the supplied environment in all datastores
				fmt.Println("\n[ INFO ] --> Deleting group: " + *group + " from Environment: " + ENV + " in provisioner............\n")
				deleteGroup(*group, ENV, "provisioner")
				fmt.Println("\n[ OK ] --> Successfully deleted group: " + *group + " from Environment: " + ENV + " in provisioner.\n")

				// Update Inventory File
				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in provisioner...............\n")
				updateInventoryFile(ENV, "provisioner")
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in provisioner.\n")

			} else {
				fmt.Println("\n[ INFO ] --> Group: " + *group + " does not exist in datastore: provisioner...skipping delete.\n")
			}

			if groupExists(*group, ENV, "custodian") {
				//delete supplied group from the supplied environment in custodian
				fmt.Println("\n[ INFO ] --> Deleting group: " + *group + " from Environment: " + ENV + "in custodian............\n")
				deleteGroup(*group, ENV, "custodian")
				fmt.Println("\n[ OK ] --> Successfully deleted group: " + *group + " from Environment: " + ENV + " in custodain.\n")

				// Update Inventory File
				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in custodian...............\n")
				updateInventoryFile(ENV, "custodian")
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in custodian.\n")

			} else {
				fmt.Println("\n[ INFO ] --> Group: " + *group + " does not exist in datastore: custodian...skipping delete.\n")
			}

			os.Exit(0)

		} else {
			if !envExists(ENV, *datastore) {
				fmt.Println("\n[ ERROR ] --> The environment: " + ENV + " does not exist in the datastore: " + *datastore + ".\n")
				os.Exit(1)
			}
			if groupExists(*group, ENV, *datastore) {
				// delete supplied group from the supplied environment in datastore
				fmt.Println("\n[ INFO ] --> Deleting group: " + *group + " from Environment: " + ENV + " in datastore: " + *datastore + "............\n")
				deleteGroup(*group, ENV, *datastore)
				fmt.Println("\n[ OK ] --> Successfully deleted group: " + *group + " from Environment: " + ENV + " in datastore: " + *datastore + ".\n")

				// Update Inventory File
				fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
				updateInventoryFile(ENV, *datastore)
				fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")
				os.Exit(0)
			} else {
				fmt.Println("\n[ ERROR ] --> Group: " + *group + " does not exist in datastore: " + *datastore + "...skipping delete.\n")
				os.Exit(1)
			}
		}

	}

	if *movehost {
		//ensure necessary sub-flag values were supplied
		if *fqdn == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -fqdn when it is required.\n")
			os.Exit(1)
		}

		if *togroup == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -to-group when it is required.\n")
			os.Exit(1)
		}

		if *fromgroup == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -from-group when it is required.\n")
			os.Exit(1)
		}

		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -datastore when it is required.\n")
			os.Exit(1)
		}

		// validate environment
		if !envExists(ENV, *datastore) {
			fmt.Println("\n[ ERROR ] --> The Environment: " + ENV + " does not exist in the database.\n")
			os.Exit(1)
		}

		// validate host
		if !hostExists(*fqdn, ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The Host: " + *fqdn + " does not exist in Environment: " + ENV + ".\n")
			os.Exit(1)
		}

		// validate group to which the host will be copied
		if !groupExists(*togroup, ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The Group: " + *togroup + " does not exist in Environment: " + ENV + ".\n")
			os.Exit(1)
		}

		// validate group from which the host will be removed
		if !groupExists(*fromgroup, ENV, *datastore) {
			fmt.Println("\n[ FAILED ] --> The Group: " + *fromgroup + " does not exist in Environment: " + ENV + ".\n")
			os.Exit(1)
		}

		// detach supplied host from the supplied group
		fmt.Println("\nDetaching host: " + *fqdn + " from group: " + *fromgroup + "............\n")
		detachGroupFromHost(*fqdn, *fromgroup, ENV, *datastore)
		detachHostFromGroup(*fqdn, *fromgroup, ENV, *datastore)
		fmt.Println("\n[ OK] --> Successfully detached host: " + *fqdn + " from group: " + *fromgroup + "............\n")

		// attach supplied host to the requested group
		attachHost(*fqdn, *togroup, ENV, *datastore)

		// Update Inventory File
		fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in " + *datastore + "...............\n")
		updateInventoryFile(ENV, *datastore)
		fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in " + *datastore + ".\n")

		os.Exit(0)

	}

	if *addenv {
		// ensure that the data store has been provided
		if *datastore == "EMPTY" {
			fmt.Println("\n[ ERROR ] --> You did not provide a value for the sub-flag -dataStore when it is required.\n")
			os.Exit(1)
		}

		if envExists(*environment, *datastore) {
			fmt.Println("\n[ FAILED ] --> The environment: " + *environment + " already Exists in the database.\n")
			os.Exit(1)
		}

		// create prefix
		ePrefix := strings.ToLower(strings.Replace(*environment, "-", "_", -1))

		// create AnsibleEnvironment struct and populate fields
		anEnvironment := new(AnsibleEnvironment)
		anEnvironment.Prefix = ePrefix
		anEnvironment.Name = *environment

		if *datastore == "all" {
			addEnvironment(anEnvironment, "provisioner")
			fmt.Println("\n[ INFO ] --> Creating Inventory file for Environment: " + ENV + " in provisioner.......\n")
			// add Inventory file
			createInventoryFile(anEnvironment.Name, "provisioner")
			fmt.Println("\n[ OK ] --> Successfully Created Inventory File for " + ENV + " in provisioner.\n")

			addEnvironment(anEnvironment, "custodian")
			fmt.Println("\n[ INFO ] --> Creating Inventory file for Environment: " + ENV + " in custodian.......\n")
			// add Inventory file
			createInventoryFile(anEnvironment.Name, "custodian")
			fmt.Println("\n[ OK ] --> Successfully Created Inventory File for " + ENV + " in custodian.\n")
		} else {
			addEnvironment(anEnvironment, *datastore)
			fmt.Println("\n[ INFO ] --> Creating Inventory file for Environment: " + ENV + "...............\n")
			// add Inventory file
			createInventoryFile(anEnvironment.Name, *datastore)
			fmt.Println("\n[ OK ] --> Successfully Created Inventory File for " + ENV + ".\n")
		}

		os.Exit(0)
	}

}

func listInventory() {
	envGroupsDb := strings.ToLower(strings.Replace(ENV, "-", "_", -1)) + "_groups"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB("provisioner").C(envGroupsDb)

	// get Iterator of items in the collection
	iter := c.Find(nil).Iter()
	var ansibleGrps AnsibleGroups
	groupsSlice := make(map[string][]string)
	for iter.Next(&ansibleGrps) {
		for k := range ansibleGrps.Members {
			groupsSlice[k] = ansibleGrps.Members[k]
		}
	}

	// convert groupsSlice map to nicely formated json
	b, err := json.MarshalIndent(groupsSlice, "", "   ")
	if err != nil {
		fmt.Println("error:", err)
	}

	// print json group document
	os.Stdout.Write(b)
}

func listHostVars() {
	varMap := make(map[string]string)
	b, err := json.Marshal(varMap)

	if err != nil {
		fmt.Println("error: ", err)
	}

	os.Stdout.Write(b)

}

func addHost(newHost AnsibleHost, database string) {
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// set up hosts collection reference
	hCollection := strings.ToLower(strings.Replace(newHost.Environment, "-", "_", -1)) + "_hosts"

	// attatch session to desired database and collection
	c := session.DB(database).C(hCollection)

	fmt.Println("\n[ INFO ] --> Adding " + newHost.Fqdn + " to Environment: " + newHost.Environment + " in database: " + database + "......\n")
	err = c.Insert(&newHost)
	if err != nil {
		fmt.Println("\n[ ERROR ] -- Failed to add host to database: " + database + ".\n")
		os.Exit(1)
	}

	fmt.Println("\n[ OK ] --> Successfully added " + newHost.Fqdn + " to " + newHost.Environment + "\n")

	// attach new host to default environment _all group
	allGroup := strings.ToLower(strings.Replace(newHost.Environment, "-", "_", -1)) + "_all"
	attachHost(newHost.Fqdn, allGroup, newHost.Environment, database)

}

func addPulpClient(pulpClient PulpClient, database string) {
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	// attatch session to desired database and collection
	c := session.DB(database).C("pulp_clients")

	fmt.Println("\nAdding " + pulpClient.Fqdn + " to the Pulp Client Database in " + database + "......\n")
	err = c.Insert(&pulpClient)
	if err != nil {
		fmt.Println("\n[ ERROR ] -- Failed to add Pulp client to the " + database + " database.\n")
		os.Exit(1)
	}
}

func addEnvironment(newEnv *AnsibleEnvironment, database string) {
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C("environments")

	fmt.Println("\nAdding Environment " + newEnv.Name + " to Inventory........")
	err = c.Insert(&newEnv)
	if err != nil {
		fmt.Println("\n[ ERROR ] -- Failed to add Environment to database.\n")
		panic(err)
	}
	fmt.Println("\n[ OK ] -- successfully added Environment: " + newEnv.Name + "\n")

	// set up group session
	allGroup := AnsibleGroups{}
	allGroup.Name = strings.ToLower(strings.Replace(newEnv.Name, "-", "_", -1)) + "_all"

	// setup the group members map with empty members slice
	groupMembers := map[string][]string{allGroup.Name: make([]string, 0)}
	allGroup.Members = groupMembers
	allGroup.Description = "Default Group for all members in " + newEnv.Name
	allGroup.Environment = newEnv.Name

	// Add the group the requested environment
	addGroup(allGroup, database)

}

func listGroups(ansibleEnv, database string) {

	// Set up groups collection
	groupsCollection := strings.ToLower(strings.Replace(ansibleEnv, "-", "_", -1)) + "_groups"
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(groupsCollection)

	// header
	fmt.Println("\n--BEGIN--\n")
	fmt.Println("\n=====================   [ " + ansibleEnv + " ]   =====================\n")

	// get Iterator of groups in the environment groups collection
	iter := c.Find(nil).Iter()
	var ansibleGrps AnsibleGroups
	for iter.Next(&ansibleGrps) {
		// Print out each group name and its description
		fmt.Println("\n| Groupname: " + ansibleGrps.Name)
		fmt.Println("| Description: " + ansibleGrps.Description + "\n|")
	}

	// footer
	fmt.Println("\n\n\n--END--\n")

}

func addGroup(newGroup AnsibleGroups, database string) {
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// set up groups collection reference
	gCollection := strings.ToLower(strings.Replace(ENV, "-", "_", -1)) + "_groups"

	// attatch session to desired database and collection
	c := session.DB(database).C(gCollection)

	// add group to database
	fmt.Println("\n[ INFO ] --> Adding Group " + newGroup.Name + " to the Inventory Environment " + ENV + " in datastore: " + database + "......\n")
	err = c.Insert(&newGroup)

	// check for errors
	if err != nil {
		fmt.Println("\n[ ERROR ] -- Failed to add group to database: " + database + ".\n")
		panic(err)
	}
	fmt.Println("\n[ OK ] --> Successfully added group: " + newGroup.Name + " to the environment: " + ENV + ".\n")

	fmt.Println("\n[ INFO ] --> Associating group: " + newGroup.Name + " to Environment: " + ENV + " in datastore: " + database + ".\n")
	assocGroupToEnv(newGroup.Name, ENV, database)
	fmt.Println("\n[ OK ] -- Successfully associated group: " + newGroup.Name + " to Environment: " + ENV + " in datastore: " + database + "\n")
}

func cloneHost(templateName, hostName, envName, database string) {

	// setup db prefix
	envDbPrefix := strings.ToLower(strings.Replace(envName, "-", "_", -1))

	// setup host collection reference
	envHostsCollection := envDbPrefix + "_hosts"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(envHostsCollection)

	result := AnsibleHost{}
	// executes the query and returns single match
	err = c.Find(bson.M{"fqdn": templateName}).One(&result)
	if err != nil {
		panic(err)
	}

	// copy template host field values to new host except Fqdn field value
	newHost := AnsibleHost{}
	newHost.Groups = result.Groups
	newHost.Environment = result.Environment
	newHost.Fqdn = hostName

	// add New Host to database
	fmt.Println("\n[ INFO ] --> Creating New host: " + newHost.Fqdn + " from Template host: " + templateName + "\n")
	err = c.Insert(&newHost)
	if err != nil {
		panic(err)
	}

	fmt.Println("\n[ OK ] --> Successfully created " + newHost.Fqdn + " from Template: " + templateName + "\n")

	for k := range newHost.Groups {
		fmt.Println("\nAttaching " + hostName + " to group: " + k + "\n")
		assocHostToGroup(hostName, k, envDbPrefix, database)
		fmt.Println("\n[ OK ] --> Successfully attached " + hostName + " to group: " + k + "\n")

	}

	clonePulpClient(templateName, hostName, database)

}

func clonePulpClient(template, host, database string) {
	// setup host collection reference
	pulpCollection := "pulp_clients"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(pulpCollection)

	result := PulpClient{}
	// executes the query and returns single match
	err = c.Find(bson.M{"fqdn": template}).One(&result)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Could not find template host: " + template + " in " + database + ".\n")
		os.Exit(1)
	}

	result.Fqdn = host

	err = c.Insert(&result)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Unable to add host: " + host + " to pulp client database in " + database + ".\n")
		os.Exit(1)
	}

}

func attachHost(hostName string, groupName string, envName string, database string) {
	// Set up Prefix and host collection details
	envDbPrefix := strings.ToLower(strings.Replace(envName, "-", "_", -1))
	envHostsCollection := envDbPrefix + "_hosts"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(envHostsCollection)

	result := AnsibleHost{}
	// executes the query and returns single match
	err = c.Find(bson.M{"fqdn": hostName}).One(&result)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to find host: " + hostName + " in database: " + database + ".....Exiting.\n")
		os.Exit(1)
	}

	// adds host to group
	fmt.Println("\nAttaching " + hostName + " to " + groupName + "............\n")

	if !result.Groups[groupName] {
		result.Groups[groupName] = true
		assocHostToGroup(hostName, groupName, envDbPrefix, database)
		assocGroupToHost(hostName, groupName, envDbPrefix, database)
	}

	fmt.Println("\n[ OK ] --> Successfully attached " + hostName + " to " + groupName + " \n")

}

func assocGroupToHost(hostName, groupName, envDbPrefix, database string) {
	hostsCollection := envDbPrefix + "_hosts"
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to connect to MongoDB at: " + MONGOIP + "\n")
		os.Exit(1)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(hostsCollection)

	result := AnsibleHost{}
	// executes the query and returns single match
	err = c.Find(bson.M{"fqdn": hostName}).One(&result)
	if err != nil {
		fmt.Println("\nFailed to find host: " + hostName + " in database: " + database + " \n")
		os.Exit(1)
	}

	if !result.Groups[groupName] {
		result.Groups[groupName] = true
	}

	err = c.Update(bson.M{"fqdn": hostName}, bson.M{"fqdn": result.Fqdn, "groups": result.Groups, "environment": result.Environment})

	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to associate group: " + groupName + " to host: " + hostName + " in database: " + database + " \n")
		os.Exit(1)
	}

	fmt.Println("\n[ OK ] --> Successfully associated group: " + groupName + " to host: " + hostName + " in database: " + database + "\n")
}

func assocHostToGroup(hostName, groupName, envDbPrefix, database string) {
	groupsCollection := envDbPrefix + "_groups"
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(groupsCollection)

	result := AnsibleGroups{}
	// executes the query and returns single match
	err = c.Find(bson.M{"name": groupName}).One(&result)
	if err != nil {
		panic(err)
	}

	//add member to updated group -- added to members group slice
	result.Members[groupName] = append(result.Members[groupName], hostName)

	err = c.Update(bson.M{"name": groupName}, bson.M{"name": result.Name, "members": result.Members, "description": result.Description, "environment": result.Environment})

	if err != nil {
		fmt.Println("[ ERROR ] --> Failed to Update Group with new host")
		os.Exit(1)
	}

}

func assocGroupToEnv(gName, gEnv, database string) {
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C("environments")

	result := AnsibleEnvironment{}
	// executes the query and returns single match
	err = c.Find(bson.M{"name": gEnv}).One(&result)
	if err != nil {
		panic(err)
	}

	// create update Environment and copy values from pre-updated Environment
	uEnv := AnsibleEnvironment{}
	uEnv.Name = result.Name
	uEnv.Prefix = result.Prefix
	uEnv.Groups = result.Groups

	// update the Groups Map to include the new Group if not already there
	if !uEnv.Groups[gName] {
		uEnv.Groups[gName] = true
	}

	// replace existing environment with the updated environment
	err = c.Update(bson.M{"name": gEnv}, bson.M{"name": uEnv.Name, "groups": uEnv.Groups, "prefix": uEnv.Prefix})

	if err != nil {
		fmt.Println("Failed to update Environment: " + gEnv + " in database: " + database + "\n")
		os.Exit(1)
	}

	fmt.Println("\n[ OK ] --> Successfully updated Environment: " + ENV + " in database: " + database + ".\n")

}

func displayHost(hName, hEnv, database string) {

	colName := strings.ToLower(strings.Replace(hEnv, "-", "_", -1)) + "_hosts"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(colName)

	result := AnsibleHost{}
	// executes the query and returns single match
	err = c.Find(bson.M{"fqdn": hName}).One(&result)
	if err != nil {
		panic(err)
	}

	fmt.Println("\n--BEGIN--\n|\n=====================   [ Details ]   =====================\n|")
	fmt.Println("| Hostname: " + result.Fqdn + "\n| Environment: " + result.Environment + "\n|\n|")
	fmt.Println("|\n=====================   [ Groups ]   ======================\n|")
	for k := range result.Groups {
		fmt.Println("| " + k)
	}
	fmt.Println("|\n|\n|\n--END--\n")

}

func displayGroup(gName, gEnv, database string) {

	colName := strings.ToLower(strings.Replace(gEnv, "-", "_", -1)) + "_groups"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(colName)

	result := AnsibleGroups{}
	// executes the query and returns single match
	err = c.Find(bson.M{"name": gName}).One(&result)
	if err != nil {
		panic(err)
	}

	fmt.Println("\n--BEGIN--\n\n=====================   [ Details ]   =====================\n|")
	fmt.Println("| Groupname: " + result.Name + "\n| Description: " + result.Description + "\n| Environment: " + result.Environment + "\n|\n|")
	fmt.Println("|\n=====================   [ Hosts ]   ======================\n|")
	memberSlice := result.Members[gName]
	for _, item := range memberSlice {
		fmt.Println("| " + item)
	}
	fmt.Println("|\n|\n\n--END--\n")

}

func listGroupOptions(ansibleEnv, database string) {
	// Set up groups collection
	groupsCollection := strings.ToLower(strings.Replace(ansibleEnv, "-", "_", -1)) + "_groups"
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(groupsCollection)

	// get Iterator of groups in the environment groups collection
	iter := c.Find(nil).Iter()
	var ansibleGrps AnsibleGroups
	for iter.Next(&ansibleGrps) {
		// Print out each group name
		fmt.Println(ansibleGrps.Name)
	}
}

func listHostOptions(ansibleEnv, database string) {
	// Set up groups collection
	hostsCollection := strings.ToLower(strings.Replace(ansibleEnv, "-", "_", -1)) + "_hosts"
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(hostsCollection)

	// get Iterator of groups in the environment groups collection
	iter := c.Find(nil).Iter()
	var ansibleHost AnsibleHost
	for iter.Next(&ansibleHost) {
		// Print out each host name
		fmt.Println(ansibleHost.Fqdn)
	}

}

func detachGroupFromHost(hostName, groupName, envName, database string) {
	// setup db prefix
	envDbPrefix := strings.ToLower(strings.Replace(envName, "-", "_", -1))

	// setup host collection reference
	envHostsCollection := envDbPrefix + "_hosts"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(envHostsCollection)

	result := AnsibleHost{}
	// executes the query and returns single match
	err = c.Find(bson.M{"fqdn": hostName}).One(&result)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to find host: " + hostName + " in " + database + "....skipping group detach.\n")
		return
	}

	// remove group from hosts groups Map
	if result.Groups[groupName] {
		delete(result.Groups, groupName)
	}

	// updating host
	err = c.Update(bson.M{"fqdn": result.Fqdn}, bson.M{"fqdn": result.Fqdn, "groups": result.Groups, "environment": result.Environment})
	if err != nil {
		panic(err)
	}
}

func detachHostFromGroup(hostName, groupName, envName, database string) {
	// setup db prefix
	envDbPrefix := strings.ToLower(strings.Replace(envName, "-", "_", -1))

	// setup Group Collection reference
	groupsCollection := envDbPrefix + "_groups"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(groupsCollection)

	result := AnsibleGroups{}
	// executes the query and returns single match
	err = c.Find(bson.M{"name": groupName}).One(&result)
	if err != nil {
		panic(err)
	}

	uGroup := AnsibleGroups{}
	uGroup.Members = result.Members
	uGroup.Description = result.Description
	uGroup.Environment = result.Environment
	uGroup.Name = result.Name

	for i, v := range uGroup.Members[groupName] {
		if v == hostName {
			gslice := uGroup.Members[groupName]
			n := i + 1
			gslice = append(gslice[:i], gslice[n:]...)
			uGroup.Members[groupName] = gslice
			break
		}
	}

	err = c.Update(bson.M{"name": groupName}, bson.M{"name": uGroup.Name, "members": uGroup.Members, "description": uGroup.Description, "environment": uGroup.Environment})

	if err != nil {
		fmt.Println("Failed to Update Group after detach")
		panic(err)
	}
}

func deleteHost(hostName, hostEnv, database string) {
	colName := strings.ToLower(strings.Replace(hostEnv, "-", "_", -1)) + "_hosts"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(colName)

	result := AnsibleHost{}
	// executes the query and returns single match
	err = c.Find(bson.M{"fqdn": hostName}).One(&result)
	if err != nil {
		fmt.Println("\n[ WARNING ] The host: " + hostName + " does not exist in " + database + ".\n")
		return
	}

	for group := range result.Groups {
		detachHostFromGroup(result.Fqdn, group, ENV, database)
	}

	err = c.Remove(bson.M{"fqdn": hostName})
	if err != nil {
		panic(err)
	}
}

func deletePulpClient(hostName, database string) {
	colName := "pulp_clients"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(colName)

	result := PulpClient{}
	// executes the query and returns single match
	err = c.Find(bson.M{"fqdn": hostName}).One(&result)
	if err != nil {
		fmt.Println("\n[ WARNING ] --> The host: " + hostName + " does not exist in the Pulp database. Continuing without deleting.\n")
		return
	}

	err = c.Remove(bson.M{"fqdn": hostName})
	if err != nil {
		panic(err)
	}

	fmt.Println("\n[ OK ] --> Successfully deleted host: " + hostName + " from the Pulp Repository database\n")

}

func deleteGroup(groupName, envName, database string) {

	// Set up Groups collection reference for the supplied environment
	colName := strings.ToLower(strings.Replace(envName, "-", "_", -1)) + "_groups"
	allGroup := strings.ToLower(strings.Replace(envName, "-", "_", -1)) + "_all"

	if groupName == allGroup {
		if envExists(envName, database) {
			fmt.Println("\n[ ERROR ] --> Sorry, you cannot delete the group: " + groupName + "because the environment to which this group belongs still exists.\n")
			fmt.Println("You must delete the following environment: " + envName + "first.\n")
			os.Exit(1)
		}
	}

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(colName)

	result := AnsibleGroups{}
	// executes the query and returns single match
	err = c.Find(bson.M{"name": groupName}).One(&result)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to find group: " + groupName + " in " + database + ".\n")
		os.Exit(1)
	}

	for _, v := range result.Members[groupName] {
		fmt.Println("\n[ INFO ] --> detaching group: " + groupName + " from host: " + v + ".\n")
		detachGroupFromHost(v, groupName, envName, database)
	}

	removeGroupFromEnv(groupName, ENV, database)

	err = c.Remove(bson.M{"name": result.Name})
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Unable to remove group: " + groupName + " from groups database in " + database + ".\n")
		os.Exit(1)
	}

}

func removeGroupFromEnv(groupname, environment, database string) {
	envCollection := "environments"
	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(envCollection)

	ansibleEnv := AnsibleEnvironment{}
	err = c.Find(bson.M{"name": environment}).One(&ansibleEnv)
	if err != nil {
		fmt.Println("\n[ ERROR ] -- The environment: " + environment + " does not exist in database: " + database + ".\n")
		os.Exit(1)
	}

	if ansibleEnv.Groups[groupname] {
		fmt.Println("\n[ INFO ] --> Attempting to remove Group: " + groupname + " from environment: " + environment + ".\n ")
		delete(ansibleEnv.Groups, groupname)

		// replace existing environment with the updated environment
		err = c.Update(bson.M{"name": ENV}, bson.M{"name": ENV, "groups": ansibleEnv.Groups, "prefix": ansibleEnv.Prefix})

		if err != nil {
			fmt.Println("\n[ ERROR ] --> Failed to remove group: " + groupname + " from  Environment: " + ENV + "\n")
			os.Exit(1)
		}

		fmt.Println("\n[ OK ] --> Successfully removed group: " + groupname + " from  Environment: " + environment + ".\n")
	} else {
		fmt.Println("\n[ WARNING ] --> The group: " + groupname + " was not found in Environment" + environment + "....skipping remove.\n")
	}
}

func createInventoryFile(envName, database string) {

	fileDirMap := make(map[string]string, 2)
	fileDirMap["provisioner"] = "/apps/ansible-provisioner-inventories/"
	fileDirMap["custodian"] = "/apps/ansible-inventories/"

	envDir := strings.ToLower(strings.Replace(envName, "-", "_", -1))
	invFile := InventoryFile{}
	invFile.Path = fileDirMap[database] + envDir + "/" + envDir + ".inventory"
	invFile.Environment = envName

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C("inventory_files")

	err = c.Insert(&invFile)
	if err != nil {
		fmt.Println("\n[ ERROR ] -- Failed to add inventory file to inventory files collection in " + database + ".\n")
		os.Exit(1)
	}

	// create environment inventory file directory
	inventoryDir := fileDirMap[database] + envDir
	err = os.Mkdir(inventoryDir, 0644)
	if err != nil {
		fmt.Println("\n[ ERROR ] -- Failed to create inventory directory: " + inventoryDir + ".\n")
		os.Exit(1)
	}

	// create environment inventory file backup directory
	backupDir := inventoryDir + "/backups"
	err = os.Mkdir(backupDir, 0644)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to create inventory backup directory: " + backupDir + " in " + database + ".\n")
		os.Exit(1)
	}

	// create file header
	fileHeader := "# -- !!! WARNING !!! -- This File is managed by provisioner, any changes will be over-written\n # on the next provisioner run.\n#\n#\n"

	// create the physical inventory file
	f, err := os.Create(invFile.Path)

	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to create inventory file: " + invFile.Path + " in " + database + ".\n")
		os.Exit(1)
	}

	// ensure inventory file handle is released
	defer f.Close()

	// write file header out to inventory file
	_, err = f.WriteString(fileHeader)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Falure writing header to inventory file: " + invFile.Path + " in " + database + ".\n")
		os.Exit(1)
	}

	f.Sync()
}

func updateInventoryFile(envName, database string) {

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C("inventory_files")

	resultFile := InventoryFile{}
	// executes the query and returns single match
	err = c.Find(bson.M{"environment": envName}).One(&resultFile)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> The environemnt: " + ENV + " could not be found in database: " + database + ".\n")
		os.Exit(1)
	}

	envDir := strings.ToLower(strings.Replace(envName, "-", "_", -1))
	//timeStampSlice := strings.Fields(time.Now().String())
	//timeStamp := timeStampSlice[0]
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	filePath := resultFile.Path
	backupPath := "/apps/ansible-provisioner-inventories/" + envDir + "/backups/" + envDir + ".inventory." + timeStamp
	// backup the file
	err = os.Rename(filePath, backupPath)
	if err != nil {
		panic(err)
	}

	//Set up groups collection
	gCollection := strings.ToLower(strings.Replace(envName, "-", "_", -1)) + "_groups"
	groupSession, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer groupSession.Close()
	gC := groupSession.DB(database).C(gCollection)

	fileHeader := "# -- !!! WARNING !!! -- This File is managed by provisioner, any changes will be over-written\n# on the next provisioner run.\n#\n#\n"

	// create new inventory file
	f, err := os.Create(filePath)

	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to create inventory file for environment: " + ENV + " in database: " + database + ".\n")
		os.Exit(1)
	}

	// ensure inventory file handle is released
	defer f.Close()

	_, err = f.WriteString(fileHeader)

	// get Iterator of items in the collection
	iter := gC.Find(nil).Iter()
	var ansibleGrps AnsibleGroups
	//groupsSlice := make(map[string][]string)
	for iter.Next(&ansibleGrps) {
		// write Group Description as Comment
		_, err = f.WriteString("# " + ansibleGrps.Description + "\n")
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString("[" + ansibleGrps.Name + "]\n")
		if err != nil {
			panic(err)
		}

		for k := range ansibleGrps.Members[ansibleGrps.Name] {
			_, err = f.WriteString(ansibleGrps.Members[ansibleGrps.Name][k] + "\n")
			if err != nil {
				panic(err)
			}
		}

		_, err = f.WriteString("\n\n\n")
		if err != nil {
			panic(err)
		}
	}

	f.Sync()

}

// Host validation function
func hostExists(hostName, envName, database string) bool {

	//Set up hosts collection reference
	hCollection := strings.ToLower(strings.Replace(envName, "-", "_", -1)) + "_hosts"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(hCollection)

	doesExist := false

	// get Iterator of items in the collection
	iter := c.Find(nil).Iter()
	var ansibleHost AnsibleHost
	for iter.Next(&ansibleHost) {
		if ansibleHost.Fqdn == hostName {
			doesExist = true
		}
	}

	return doesExist

}

// Group validation function
func groupExists(groupName, envName, database string) bool {

	//Set up groups collection
	gCollection := strings.ToLower(strings.Replace(envName, "-", "_", -1)) + "_groups"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C(gCollection)

	doesExist := false

	// get Iterator of items in the collection
	iter := c.Find(nil).Iter()
	var ansibleGrp AnsibleGroups
	for iter.Next(&ansibleGrp) {
		if ansibleGrp.Name == groupName {
			doesExist = true
		}
	}

	return doesExist

}

// Environment validation function
func envExists(environment, database string) bool {

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c := session.DB(database).C("environments")

	doesExist := false

	// get Iterator of items in the collection
	iter := c.Find(nil).Iter()
	var ansibleEnv AnsibleEnvironment
	for iter.Next(&ansibleEnv) {
		if ansibleEnv.Name == environment {
			doesExist = true
		}
	}

	return doesExist
}

// pull a single host out of custodian database and put it in the provisioner database
func pullOneHost(host string) {

	provDB := "provisioner"
	custDB := "custodian"

	hostCollection := strings.ToLower(strings.Replace(ENV, "-", "_", -1)) + "_hosts"
	groupCollection := strings.ToLower(strings.Replace(ENV, "-", "_", -1)) + "_groups"
	pulpCollection := "pulp_clients"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		fmt.Println("\n[ FAILED ] --> Unable to obtain a connection to MongoDB.\n")
		os.Exit(1)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c1 := session.DB(custDB).C(hostCollection)

	custHost := AnsibleHost{}
	// executes the query and returns single match
	err = c1.Find(bson.M{"fqdn": host}).One(&custHost)
	if err != nil {
		fmt.Println("\n[ FAILED ] --> The host: " + host + " does not exist in the custodian database.\n")
		os.Exit(1)
	}

	// copy host into provisioner database
	c2 := session.DB(provDB).C(hostCollection)

	fmt.Println("\nAdding " + custHost.Fqdn + " to Provisioner Environment: " + ENV + "......\n")
	err = c2.Insert(&custHost)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to add host: " + host + " to provisioner database.\n")
		os.Exit(1)
	}

	// attach session to provisioner groups collection to attach host to the appropriate groups
	c3 := session.DB(provDB).C(groupCollection)

	for group := range custHost.Groups {
		// ensure that a provisioner group exists for each group that exists in custodian

		provGroup := AnsibleGroups{}
		// executes the query and returns single match
		err = c3.Find(bson.M{"name": group}).One(&provGroup)
		if err != nil {
			fmt.Println("\n[ INFO ] --> Failed to locate local provisioner group named: " + group + "... validating group.\n")
			custGroup := AnsibleGroups{}
			c4 := session.DB(custDB).C(groupCollection)
			err = c4.Find(bson.M{"name": group}).One(&custGroup)
			if err != nil {
				fmt.Println("\n[ ERROR ] --> Group: " + group + " is missing from custodian but is referenced in host: " + host + ".\n")
				fmt.Println("[ ERROR ] --> Something is out of sync between custodian and provisioner\n")
				fmt.Println("[ FATAL ERROR ] --> Manual Intervention is required.\n")
				os.Exit(1)
			}
			fmt.Println("\n[ INFO ] --> Successfully validated group: " + group + " in custodian...continuing with add.\n")

			// add group to database
			provGroup := AnsibleGroups{}
			provGroup.Members = map[string][]string{group: make([]string, 0)}
			provGroup.Description = custGroup.Description
			provGroup.Environment = custGroup.Environment
			provGroup.Name = custGroup.Name
			fmt.Println("\n[ INFO ] --> Adding group: " + provGroup.Name + " to the provisioner database.\n")
			c5 := session.DB(provDB).C(groupCollection)
			err = c5.Insert(&provGroup)

			// check for errors
			if err != nil {
				fmt.Println("\n[ ERROR ] --> Failed to add group: " + group + " to provisioner database.\n")
				os.Exit(1)
			}
			// attach the group the corresponding provisioner environment
			fmt.Println("\n[ INFO ] --> Adding Group " + group + " to the Provisioner Inventory Environment " + ENV + "......\n")
			assocGroupToEnv(group, ENV, "provisioner")
			fmt.Println("\n[ OK ] -- Successfully added group to " + ENV + "\n")
		}

		//add member to updated group -- added to members group slice
		isPresent := false
		for member := range provGroup.Members[group] {
			if provGroup.Members[group][member] == host {
				isPresent = true
			}
		}

		if !isPresent {
			provGroup.Members[group] = append(provGroup.Members[group], host)
		} else {
			fmt.Println("\n[ INFO ] --> The host: " + host + " already exists in group: " + provGroup.Name + "... skipping add.\n")
		}

		c6 := session.DB(provDB).C(groupCollection)
		err = c6.Update(bson.M{"name": provGroup.Name}, bson.M{"name": provGroup.Name, "members": provGroup.Members, "description": provGroup.Description, "environment": provGroup.Environment})

		if err != nil {
			fmt.Println("\n[ ERROR ] --> Failed to Update Group: " + group + " with new host: " + host + ".\n")
			os.Exit(1)
		}
	}

	// delete host from custodian database
	// attatch session to desired database and collection
	c7 := session.DB(custDB).C(hostCollection)

	delHost := AnsibleHost{}
	// executes the query and returns single match
	err = c7.Find(bson.M{"fqdn": host}).One(&delHost)
	if err != nil {
		fmt.Println("The host: " + host + " does not exist in custodian and therefore will not be deleted.\n")
	}

	for group := range delHost.Groups {

		// attatch session to desired database and collection
		c8 := session.DB(custDB).C(groupCollection)

		resultGroup := AnsibleGroups{}
		// executes the query and returns single match
		err = c8.Find(bson.M{"name": group}).One(&resultGroup)
		if err != nil {
			fmt.Println("\n[ ERROR ] --> Could not find group: " + group + " in custodian database when it should be present.\n")
			os.Exit(1)
		}

		for i, v := range resultGroup.Members[group] {
			if v == host {
				gslice := resultGroup.Members[group]
				n := i + 1
				gslice = append(gslice[:i], gslice[n:]...)
				resultGroup.Members[group] = gslice
				break
			}
		}

		err = c8.Update(bson.M{"name": group}, bson.M{"name": resultGroup.Name, "members": resultGroup.Members, "description": resultGroup.Description, "environment": resultGroup.Environment})
		if err != nil {
			fmt.Println("\n[ ERROR ] --> Failed to remove host: " + host + " from group: " + group + " in custodian database.\n")
			os.Exit(1)
		}

	}

	// Delete the host from the custodian database
	c9 := session.DB(custDB).C(hostCollection)
	err = c9.Remove(bson.M{"fqdn": host})
	if err != nil {
		fmt.Println("[ ERROR ] --> Failed to remove host from custodian database.\n")
		os.Exit(1)
	}

	// Get the pulp client information from custodian and add it to provisioner
	c10 := session.DB(custDB).C(pulpCollection)
	pClient := PulpClient{}
	err = c10.Find(bson.M{"fqdn": host}).One(&pClient)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to retreive pulp information for host: " + host + " from custodian database.\n")
		os.Exit(1)
	}

	c11 := session.DB(provDB).C(pulpCollection)
	provPulpClient := PulpClient{}
	err = c11.Find(bson.M{"fqdn": host}).One(&provPulpClient)
	if err != nil {
		// Add pulp client to provisioner  pulp clients database
		addPulpClient(pClient, "provisioner")
	} else {
		fmt.Println("\n[ INFO ] --> The host: " + host + " is already in the provisioner pulp client database...skipping add.\n")
	}

	// Delete pulp client from custodian database
	fmt.Println("\n[ INFO ] --> Deleting host from custodian pulp clients database.\n")
	c12 := session.DB(custDB).C(pulpCollection)
	err = c12.Remove(bson.M{"fqdn": host})
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to remove host: " + host + " from custodian pulp clients database.\n")
		os.Exit(1)
	}
	fmt.Println("\n[ OK ] --> Successfully deleted host: " + host + " from custodain pulp clients database.\n")
}

// push hosts from provisioner database into custodian database
func pushOneHost(host string) {

	provDB := "provisioner"
	custDB := "custodian"

	hostCollection := strings.ToLower(strings.Replace(ENV, "-", "_", -1)) + "_hosts"
	groupCollection := strings.ToLower(strings.Replace(ENV, "-", "_", -1)) + "_groups"
	pulpCollection := "pulp_clients"

	// Set up connection to database server
	session, err := mgo.Dial(MONGOIP)
	if err != nil {
		fmt.Println("\n[ FAILED ] --> Unable to obtain a connection to MongoDB.\n")
		os.Exit(1)
	}
	defer session.Close()

	// attatch session to desired database and collection
	c1 := session.DB(provDB).C(hostCollection)

	provHost := AnsibleHost{}
	// executes the query and returns single match
	err = c1.Find(bson.M{"fqdn": host}).One(&provHost)
	if err != nil {
		fmt.Println("\n[ FAILED ] --> The host: " + host + " does not exist in the provisioner database.\n")
		os.Exit(1)
	}

	// copy host into custodian database
	c2 := session.DB(custDB).C(hostCollection)

	fmt.Println("\nAdding " + provHost.Fqdn + " to custodian Environment: " + ENV + "......\n")
	err = c2.Insert(&provHost)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to add host: " + host + " to provisioner database.\n")
		os.Exit(1)
	}

	// attach session to custodian groups collection to attach host to the appropriate groups
	c3 := session.DB(custDB).C(groupCollection)

	for group := range provHost.Groups {
		// ensure that a custodian group exists for each group that exists in provisioner

		custGroup := AnsibleGroups{}
		// executes the query and returns single match
		err = c3.Find(bson.M{"name": group}).One(&custGroup)
		if err != nil {
			fmt.Println("\n[ INFO ] --> Failed to locate local custodian group named: " + group + "... validating group.\n")
			provGroup := AnsibleGroups{}
			c4 := session.DB(provDB).C(groupCollection)
			err = c4.Find(bson.M{"name": group}).One(&provGroup)
			if err != nil {
				fmt.Println("\n[ ERROR ] --> Group: " + group + " is missing from provisioner but is referenced in host: " + host + ".\n")
				fmt.Println("[ ERROR ] --> Something is out of sync between custodian and provisioner\n")
				fmt.Println("[ FATAL ERROR ] --> Manual Intervention is required.\n")
				os.Exit(1)
			}
			fmt.Println("\n[ OK ] --> Successfully validated group: " + group + " in provisioner...continuing with add.\n")

			// add group to custodian database
			custGroup = provGroup
			custGroup.Members = make(map[string][]string, 0)
			custGroup.Members[group] = make([]string, 0)

			fmt.Println("\n[ INFO ] --> Adding group: " + custGroup.Name + " to the custodian database.\n")
			c5 := session.DB(custDB).C(groupCollection)
			err = c5.Insert(&custGroup)

			// check for errors
			if err != nil {
				fmt.Println("\n[ ERROR ] --> Failed to add group: " + group + " to custodian database.\n")
				os.Exit(1)
			}
			// attach the group the corresponding custodian environment
			fmt.Println("\n[ INFO ] --> Adding Group " + group + " to the Custodian Inventory Environment " + ENV + "......\n")

			//NEED TO ACTUALLY ADD THE CODE HERE FOR OTHER DB GROUP-ENV ASSOCIATION -- Associate group to env in custodian
			c6 := session.DB(custDB).C("environments")
			custEnv := AnsibleEnvironment{}
			c6.Find(bson.M{"name": ENV}).One(&custEnv)
			if !custEnv.Groups[group] {

				// update the Groups Map to include the new Group if not already there

				custEnv.Groups[group] = true

				// replace existing environment with the updated environment
				err = c6.Update(bson.M{"name": ENV}, bson.M{"name": ENV, "groups": custEnv.Groups, "prefix": custEnv.Prefix})

				if err != nil {
					fmt.Println("Failed to update Environment: " + ENV + "\n")
					os.Exit(1)
				}

				fmt.Println("\n[ OK ] --> Successfully updated Environment: " + ENV + "\n")

			}
			fmt.Println("\n[ OK ] -- Successfully added group to " + ENV + "\n")
		}
		//add member to updated group -- added to members group slice
		isPresent := false
		for member := range custGroup.Members[group] {
			if custGroup.Members[group][member] == host {
				isPresent = true
			}
		}

		if !isPresent {
			fmt.Println("\n[ INFO ] --> adding host: " + host + " to group: " + group + ".\n")
			custGroup.Members[group] = append(custGroup.Members[group], host)
		} else {
			fmt.Println("\n[ INFO ] --> The host: " + host + " already exists in group: " + custGroup.Name + "... skipping add.\n")
		}

		c7 := session.DB(custDB).C(groupCollection)
		err = c7.Update(bson.M{"name": custGroup.Name}, bson.M{"name": custGroup.Name, "members": custGroup.Members, "description": custGroup.Description, "environment": custGroup.Environment})

		if err != nil {
			fmt.Println("\n[ ERROR ] --> Failed to Update Group: " + group + " with new host: " + host + ".\n")
			os.Exit(1)
		}
	}

	// delete host from provisioner database
	// attatch session to desired database and collection
	c8 := session.DB(provDB).C(hostCollection)

	delHost := AnsibleHost{}
	// executes the query and returns single match
	err = c8.Find(bson.M{"fqdn": host}).One(&delHost)
	if err != nil {
		fmt.Println("The host: " + host + " does not exist in provisioner and therefore will not be deleted.\n")
	}

	for group := range delHost.Groups {

		// attatch session to desired database and collection
		c9 := session.DB(provDB).C(groupCollection)

		resultGroup := AnsibleGroups{}
		// executes the query and returns single match
		err = c9.Find(bson.M{"name": group}).One(&resultGroup)
		if err != nil {
			fmt.Println("\n[ ERROR ] --> Could not find group: " + group + " in provisioner database when it should be present.\n")
			os.Exit(1)
		}

		// MAY NEED ADDITIONAL BOUNDS CHECKING FOR WHEN MEMBERS SLICE LENGTH IS LESS THAN OR EQUAL TO 1
		for i, v := range resultGroup.Members[group] {
			if v == host {
				gslice := resultGroup.Members[group]
				n := i + 1
				gslice = append(gslice[:i], gslice[n:]...)
				resultGroup.Members[group] = gslice
				break
			}
		}

		err = c9.Update(bson.M{"name": group}, bson.M{"name": resultGroup.Name, "members": resultGroup.Members, "description": resultGroup.Description, "environment": resultGroup.Environment})
		if err != nil {
			fmt.Println("\n[ ERROR ] --> Failed to remove host: " + host + " from group: " + group + " in provisioner database.\n")
			os.Exit(1)
		}

	}

	// Delete the host from the provisioner database
	c10 := session.DB(provDB).C(hostCollection)
	err = c10.Remove(bson.M{"fqdn": host})
	if err != nil {
		fmt.Println("[ ERROR ] --> Failed to remove host from provisioner database.\n")
		os.Exit(1)
	}

	// Get the pulp client information from provisioner and add it to custodian
	c11 := session.DB(provDB).C(pulpCollection)
	pClient := PulpClient{}
	err = c11.Find(bson.M{"fqdn": host}).One(&pClient)
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to retreive pulp information for host: " + host + " from provisioner database.\n")
		os.Exit(1)
	}

	custPulpClient := PulpClient{}
	c12 := session.DB(custDB).C(pulpCollection)
	err = c12.Find(bson.M{"fqdn": host}).One(&custPulpClient)
	if err != nil {
		// Add pulp client to custodian pulp clients database
		fmt.Println("\n[ INFO ] --> Adding host: " + host + " to custodian pulp client database.\n")
		err = c12.Insert(&pClient)
		if err != nil {
			fmt.Println("\n[ ERROR ] --> Failed to add host: " + host + " to custodian pulp client database.\n")
			os.Exit(1)
		}

	} else {
		fmt.Println("\n[ INFO ] --> The host: " + host + " is already in the custodian pulp client database...skipping add.\n")
	}

	// Delete pulp client from provisioner database
	fmt.Println("\n[ INFO ] --> Deleting host from provisioner pulp clients database.\n")
	c13 := session.DB(provDB).C(pulpCollection)
	err = c13.Remove(bson.M{"fqdn": host})
	if err != nil {
		fmt.Println("\n[ ERROR ] --> Failed to remove host: " + host + " from provisioner pulp client database.\n")
		os.Exit(1)
	}
	fmt.Println("\n[ OK ] --> Successfully deleted host: " + host + " from provisioner pulp client database.\n")

	fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in provisioner...............\n")
	// Update Inventory File
	updateInventoryFile(ENV, "provisioner")
	fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in provisioner.\n")

	fmt.Println("\n[ INFO ] --> Updating Inventory file for Environment: " + ENV + " in custodian...............\n")
	// Update Inventory File
	updateInventoryFile(ENV, "custodian")
	fmt.Println("\n[ OK ] --> Successfully Updated Inventory File for " + ENV + " in custodian.\n")

}
