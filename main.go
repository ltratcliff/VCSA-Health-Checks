/*
Copyright (c) 2017 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
This example program shows how the `view` and `property` packages can
be used to navigate a vSphere inventory structure using govmomi.
*/

package main

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/mo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
	"log"
	"net/url"
	"os"
	"time"
)

type Inventory struct {
	VCSA   []VCSA   `yaml:"vcsa"`
	Emails []string `yaml:"emails"`
}

type VCSA struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type Host struct {
	IP          string `json:"ip"`
	Online      bool   `json:"online"`
	UsedCPU     int    `json:"used_cpu,omitempty"`
	TotalCPU    int    `json:"total_cpu,omitempty"`
	FreeCPU     int    `json:"free_cpu,omitempty"`
	UsedMemory  int    `json:"used_memory,omitempty"`
	TotalMemory int    `json:"total_memory,omitempty"`
	FreeMemory  int    `json:"free_memory,omitempty"`
}

type dbObject struct {
	Date time.Time       `json:"timestamp"`
	Data map[string]Host `json:"data"`
}

func main() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database("devlab").Collection("vcsa")

	var inventoryStruct Inventory
	inventoryFile, err := os.ReadFile("inventory.yaml")
	if err != nil {
		log.Printf("error reading inventory.yaml: %s\n", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(inventoryFile, &inventoryStruct)
	if err != nil {
		log.Printf("error unmarshalling inventory.yaml: %s\n", err)
		os.Exit(1)
	}

	//fmt.Printf("Inventory struct: %s\n", pretty.Sprint(inventoryStruct))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var dbObj dbObject
	dbObj.Date = time.Now()
	dbObj.Data = make(map[string]Host)

	for _, vcsa := range inventoryStruct.VCSA {
		log.Printf("connecting to vcsa: %s\n", vcsa.Name)
		server := fmt.Sprintf("https://%s:%s/sdk", vcsa.Host, vcsa.Port)
		u, _ := url.Parse(server)
		u.User = url.UserPassword(vcsa.User, vcsa.Pass)

		c, err := govmomi.NewClient(ctx, u, true)
		if err != nil {
			log.Printf("error connecting to vcsa: %s\n", err)
			dbObj.Data[vcsa.Name] = Host{
				IP:     vcsa.Host,
				Online: false,
			}
			continue
		}
		defer func(c *govmomi.Client, ctx context.Context) {
			err := c.Logout(ctx)
			if err != nil {
				log.Printf("error logging out of vcsa: %s\n", err)
				os.Exit(1)
			}
		}(c, ctx)

		//// Create a view of HostSystem objects
		//var vimClient *vim25.Client = c.Client
		//m := view.NewManager(vimClient)
		//
		//v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
		//if err != nil {
		//	log.Fatalf("Failed to create container view: %v", err)
		//}
		//
		//defer v.Destroy(ctx)

		// Create a finder to search for ESXi hosts
		finder := find.NewFinder(c.Client, true)

		// Find all datacenters
		datacenters, err := finder.DatacenterList(ctx, "*")
		if err != nil {
			log.Fatalf("Failed to find datacenters: %v", err)
		}
		for _, datacenter := range datacenters {
			log.Printf("Found datacenter: %s\n", datacenter.Name())

			finder.SetDatacenter(datacenter)

			// Find all ESXi hosts
			vcsahosts, err := finder.HostSystemList(ctx, "*")
			if err != nil {
				log.Println("Error finding hosts:", err)
				return
			}

			for _, host := range vcsahosts {
				// Retrieve the host system properties, including triggered alarms
				var hs mo.HostSystem
				//err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hs)
				err = host.Properties(ctx, host.Reference(), []string{"summary"}, &hs)
				if err != nil {
					log.Printf("Error retrieving host %s: %v\n", host.Name(), err)
				}
				totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
				freeCPU := int64(totalCPU) - int64(hs.Summary.QuickStats.OverallCpuUsage)
				freeMemory := int64(hs.Summary.Hardware.MemorySize) - (int64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024)
				dbObj.Data[hs.Summary.Config.Name] = Host{
					IP:          vcsa.Host,
					Online:      true,
					UsedCPU:     int(hs.Summary.QuickStats.OverallCpuUsage),
					TotalCPU:    int(totalCPU),
					FreeCPU:     int(freeCPU),
					UsedMemory:  int(hs.Summary.QuickStats.OverallMemoryUsage),
					TotalMemory: int(hs.Summary.Hardware.MemorySize),
					FreeMemory:  int(freeMemory),
				}
			}
		}
	}

	// Convert dbObj to bson
	document := bson.D{{"timestamp", dbObj.Date}, {"data", dbObj.Data}}
	insertResult, err := collection.InsertOne(context.TODO(), document)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	// Find the document
	//var result bson.D
	//err = collection.FindOne(context.TODO(), bson.D{{Key: "name", Value: "Alice"}}).Decode(&result)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// Print the result
	// fmt.Println("Found a single document: ", result)

	// Marshal the slice to JSON
	//jsonData, err := json.Marshal(dbObj)
	//if err != nil {
	//	log.Fatalf("Error marshalling to JSON: %s", err)
	//}

	// Print the JSON data
	//fmt.Println(string(jsonData))
}
