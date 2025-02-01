package testutil

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	pb "github.com/dimitriirfan/benchmark-grpc-vs-rest-server/proto"
	"google.golang.org/protobuf/proto"
)

type JSONPerson struct {
	ID           string                 `json:"id"`
	FirstName    string                 `json:"first_name"`
	LastName     string                 `json:"last_name"`
	Email        string                 `json:"email"`
	DateOfBirth  string                 `json:"date_of_birth"`
	PhoneNumber  string                 `json:"phone_number"`
	Address      JSONAddress            `json:"address"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
	Active       bool                   `json:"active"`
	Role         string                 `json:"role"`
	ProfileImage string                 `json:"profile_image"`
	Preferences  map[string]interface{} `json:"preferences"`
}

type JSONAddress struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
}

type JSONResponse struct {
	Population []JSONPerson `json:"population"`
}

func generatePerson(index int) (*JSONPerson, *pb.Person) {
	personID := fmt.Sprintf("p%03d", index)

	if index == 1 {
		// JSON person
		jsonPerson := &JSONPerson{
			ID:          personID,
			FirstName:   "John",
			LastName:    "Smith",
			Email:       "john.smith@example.com",
			DateOfBirth: "1985-03-15T00:00:00Z",
			PhoneNumber: "+1-555-123-4567",
			Address: JSONAddress{
				Street:     "123 Main Street",
				City:       "New York",
				State:      "NY",
				Country:    "USA",
				PostalCode: "10001",
			},
			CreatedAt:    "2024-01-01T10:00:00Z",
			UpdatedAt:    "2024-01-01T10:00:00Z",
			Active:       true,
			Role:         "user",
			ProfileImage: fmt.Sprintf("https://example.com/profiles/%s.jpg", personID),
			Preferences: map[string]interface{}{
				"theme":         "dark",
				"notifications": true,
				"language":      "en",
			},
		}

		// Protobuf person
		pbPerson := &pb.Person{
			Id:          personID,
			FirstName:   "John",
			LastName:    "Smith",
			Email:       "john.smith@example.com",
			DateOfBirth: "1985-03-15T00:00:00Z",
			PhoneNumber: "+1-555-123-4567",
			Address: &pb.Address{
				Street:     "123 Main Street",
				City:       "New York",
				State:      "NY",
				Country:    "USA",
				PostalCode: "10001",
			},
			CreatedAt:    "2024-01-01T10:00:00Z",
			UpdatedAt:    "2024-01-01T10:00:00Z",
			Active:       true,
			Role:         "user",
			ProfileImage: fmt.Sprintf("https://example.com/profiles/%s.jpg", personID),
			Preferences:  make(map[string]*pb.Value),
		}

		pbPerson.Preferences["theme"] = &pb.Value{
			Kind: &pb.Value_StringValue{StringValue: "dark"},
		}
		pbPerson.Preferences["notifications"] = &pb.Value{
			Kind: &pb.Value_BoolValue{BoolValue: true},
		}
		pbPerson.Preferences["language"] = &pb.Value{
			Kind: &pb.Value_StringValue{StringValue: "en"},
		}

		return jsonPerson, pbPerson
	}

	// Sample data
	firstNames := []string{"Alex", "Jordan", "Taylor", "Morgan", "Casey", "Riley", "Sam", "Drew", "Avery", "Quinn"}
	lastNames := []string{"White", "Miller", "Moore", "Jackson", "Martin", "Lee", "Perez", "Walker", "Hall", "Young"}
	cities := []string{"Seattle", "Portland", "San Francisco", "Los Angeles", "Denver", "Chicago", "Boston", "New York"}
	states := []string{"WA", "OR", "CA", "CO", "IL", "MA", "NY", "FL", "GA", "TX", "AZ", "NV"}
	themes := []string{"light", "dark", "system"}
	roles := []string{"user", "manager", "admin"}
	languages := []string{"en", "es"}
	streets := []string{"Oak", "Maple", "Pine", "Cedar", "Elm"}
	streetTypes := []string{"Street", "Avenue", "Road", "Lane", "Drive"}

	// Generate random data
	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]
	theme := themes[rand.Intn(len(themes))]
	notifications := rand.Float32() > 0.3 // 70% chance of notifications enabled
	language := languages[rand.Intn(len(languages))]

	// Generate dates
	startDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	randomDays := rand.Intn(int(endDate.Sub(startDate).Hours() / 24))
	dob := startDate.Add(time.Duration(randomDays) * 24 * time.Hour)
	createdAt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC).Add(time.Duration(index-1) * 5 * time.Minute)

	// Create JSON person
	jsonPerson := &JSONPerson{
		ID:          personID,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       fmt.Sprintf("%s.%s@example.com", firstName, lastName),
		DateOfBirth: dob.Format(time.RFC3339),
		PhoneNumber: fmt.Sprintf("+1-555-%03d-%04d", rand.Intn(1000), rand.Intn(10000)),
		Address: JSONAddress{
			Street:     fmt.Sprintf("%d %s %s", rand.Intn(900)+100, streets[rand.Intn(len(streets))], streetTypes[rand.Intn(len(streetTypes))]),
			City:       cities[rand.Intn(len(cities))],
			State:      states[rand.Intn(len(states))],
			Country:    "USA",
			PostalCode: fmt.Sprintf("%05d", rand.Intn(90000)+10000),
		},
		CreatedAt:    createdAt.Format(time.RFC3339),
		UpdatedAt:    createdAt.Format(time.RFC3339),
		Active:       rand.Float32() > 0.1, // 90% chance of being active
		Role:         roles[rand.Intn(len(roles))],
		ProfileImage: fmt.Sprintf("https://example.com/profiles/%s.jpg", personID),
		Preferences: map[string]interface{}{
			"theme":         theme,
			"notifications": notifications,
			"language":      language,
		},
	}

	// Create Protobuf person
	pbPerson := &pb.Person{
		Id:          personID,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       jsonPerson.Email,
		DateOfBirth: jsonPerson.DateOfBirth,
		PhoneNumber: jsonPerson.PhoneNumber,
		Address: &pb.Address{
			Street:     jsonPerson.Address.Street,
			City:       jsonPerson.Address.City,
			State:      jsonPerson.Address.State,
			Country:    jsonPerson.Address.Country,
			PostalCode: jsonPerson.Address.PostalCode,
		},
		CreatedAt:    jsonPerson.CreatedAt,
		UpdatedAt:    jsonPerson.UpdatedAt,
		Active:       jsonPerson.Active,
		Role:         jsonPerson.Role,
		ProfileImage: jsonPerson.ProfileImage,
		Preferences:  make(map[string]*pb.Value),
	}

	// Set preferences in protobuf
	pbPerson.Preferences["theme"] = &pb.Value{
		Kind: &pb.Value_StringValue{StringValue: theme},
	}
	pbPerson.Preferences["notifications"] = &pb.Value{
		Kind: &pb.Value_BoolValue{BoolValue: notifications},
	}
	pbPerson.Preferences["language"] = &pb.Value{
		Kind: &pb.Value_StringValue{StringValue: language},
	}

	return jsonPerson, pbPerson
}

func GenerateFixtures(sizes []int) {
	rand.Seed(time.Now().UnixNano())

	for _, size := range sizes {
		// Generate all entries
		jsonPopulation := make([]JSONPerson, 0, size)
		pbPopulation := &pb.GetPopulationResponse{
			Population: make([]*pb.Person, 0, size),
		}

		for i := 1; i <= size; i++ {
			jsonPerson, pbPerson := generatePerson(i)
			jsonPopulation = append(jsonPopulation, *jsonPerson)
			pbPopulation.Population = append(pbPopulation.Population, pbPerson)
		}

		// Create and write JSON output
		jsonOutput := JSONResponse{Population: jsonPopulation}
		jsonData, err := json.MarshalIndent(jsonOutput, "", "  ")
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		if err := os.WriteFile(fmt.Sprintf("testutil/fixtures/fixtures_population_%d.json", size), jsonData, 0644); err != nil {
			log.Fatalf("Error writing JSON file: %v", err)
		}

		// Write protobuf binary output
		pbData, err := proto.Marshal(pbPopulation)
		if err != nil {
			log.Fatalf("Error marshaling protobuf: %v", err)
		}
		if err := os.WriteFile(fmt.Sprintf("testutil/fixtures/fixtures_population_%d.pb", size), pbData, 0644); err != nil {
			log.Fatalf("Error writing protobuf file: %v", err)
		}

		// Log sizes for comparison
		jsonData, _ = json.Marshal(jsonOutput)
		pbData, _ = proto.Marshal(pbPopulation)
		log.Printf("JSON size: %d bytes", len(jsonData))
		log.Printf("Protobuf size: %d bytes", len(pbData))
	}
}
