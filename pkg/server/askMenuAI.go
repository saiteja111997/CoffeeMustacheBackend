package server

import (
	"bytes"
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
)

type AIRequest struct {
	Query  string `json:"query"`
	CafeID uint   `json:"cafe_id"`
}

// API Handler
func (s *Server) AskMenuAI(c *fiber.Ctx) error {

	var aiRequest AIRequest
	if err := c.BodyParser(&aiRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	userQuery := aiRequest.Query
	if userQuery == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query is required"})
	}

	// Fetch categories dynamically based on the provided cafe_id
	var categories []string
	err := s.Db.Raw(`
		SELECT DISTINCT category FROM menu_items WHERE cafe_id = ?
	`, aiRequest.CafeID).Pluck("category", &categories).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch categories"})
	}

	// Convert categories to a comma-separated string
	categoryList := strings.Join(categories, ", ")

	// Dynamically include user query in the AI prompt
	var items []structures.MenuItem
	err = s.Db.Raw(`
		SELECT name FROM menu_items WHERE cafe_id = ?
	`, aiRequest.CafeID).Scan(&items).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch items"})
	}

	prompt := fmt.Sprintf(`
	You are an AI trained to generate SQL queries based on user prompts. Your goal is to accurately interpret user queries and generate SQL queries based on **intent**, not just exact keyword matches.

	### **User Query:**
	"%s"

	### **Available Categories in this Cafe:**
	%s


	### **📌 Available Categories:**
	1. Beverages
	2. Breakfast & Brunch
	3. Appetizers & Small Bites
	4. Main Course
	5. Breads & Sides
	6. Desserts & Sweets

	### **📌 Available Cuisines:**
	Italian, Mexican, Indian, Chinese, Japanese, Mediterranean, Thai, French, American, Korean, Vietnamese, Middle-Eastern, Greek, Spanish.

	### **📌 Available Dietary Labels:**
	1. Gluten-Free
	2. High-Protein
	3. Vegan
	4. Keto
	5. Lactose-Free
	6. Low-Carb
	7. Low-Fat
	8. Organic
	9. Sugar-Free
	10. Paleo
	11. Vegetarian
	12. Whole30
	13. Diabetic-Friendly

	### **📌 Available Spice Levels:**
	1. Mild
	2. Medium
	3. Spicy
	4. Extra Spicy

	### **📌 Available cm_category:**
		1. Beverages
		2. Breakfast & Brunch
		3. Appetizers & Small Bites
		4. Main Course
		5. Breads & Sides
		6. Desserts & Sweets

	### **📌 Menu Items Schema:**
	CREATE TABLE menu_items (
		id SERIAL PRIMARY KEY,
		cafe_id INT NOT NULL,
		category VARCHAR(50) NOT NULL,
		sub_category VARCHAR(50),
		name VARCHAR(100) NOT NULL,
		description TEXT,
		price DECIMAL(10,2) NOT NULL,
		is_customizable BOOLEAN DEFAULT FALSE,
		food_type VARCHAR(10) NOT NULL,
		cuisine VARCHAR(50) NOT NULL,
		dietary_labels VARCHAR(50),
		spice_level VARCHAR(20),
		ingredients TEXT,
		allergens VARCHAR(255),
		serving_size VARCHAR(50),
		calories INT,
		preparation_time INT,
		discount DECIMAL(5,2),
		popularity_score FLOAT DEFAULT 0.0,
		image_url VARCHAR(255),
		available_from VARCHAR(255),
		available_till VARCHAR(255),
		available_all_day BOOLEAN DEFAULT TRUE,
		is_available BOOLEAN DEFAULT TRUE,
		tag VARCHAR(255),
		cm_category varchar(50),
		rating FLOAT DEFAULT 0.0 NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP	
	);

	### **Instructions:**
	- Understand the **intent** behind the user's query.
	- If the query matches one of the **supported scenarios**, generate an **SQL query**.
	- Provide a **human-readable response** explaining the generated query.
	- Use proper SQL syntax, ensuring that wildcard searches use **LIKE '%%value%%'** instead of incorrect placeholders.
	- When filtering by **categories**, also consider **cm_category** in addition to category, since different cafes may have unique category names and use the categories provided in 'Available Categories in this Cafe'.
	- Remember that cm_category is a column in the menu_items table that stores the standardized category names and is fixed. Only category change from cafe to cafe.
	- Also remember to match the intents. For example user can say 'find me' or 'suggest me' or 'show me'


	### **Response Format:**
	{
		"sql": "GENERATED_SQL_QUERY_HERE",
		"response": "A response to the user prompt, e.g., 'Here are the available items under 600 that include desserts and beverages.'"
	}
	If the query does not match any supported scenario, return:
	{
		"sql": "",
		"response": "I'm only trained to help you explore menu items based on price, category, tags, cuisine, dietary preferences, spice levels, popularity, availability, customization options, and other menu-related filters."
	}

	### **✅ Supported Scenarios**

	#### ** Price-Based Queries**
	- _"Show me items below 300."_
	SELECT * FROM menu_items WHERE price < 300;
	- _"Find me dishes between 200 and 500."_
	SELECT * FROM menu_items WHERE price BETWEEN 200 AND 500;

	#### **Category-Based Queries (Including CM Category and strictly use categories provided in the input while filtering category column)**
	- _"Show me all desserts available in this cafe."_
	SELECT * FROM menu_items WHERE cm_category = 'desserts & sweets' OR category LIKE '%%dessert%%';
	- _"I want to see pizzas and pastas available here."_
	SELECT * FROM menu_items WHERE category LIKE '%%pizza%%' OR category LIKE '%%pasta%%';
	- _"Find me burgers, sandwiches, and wraps."_
	SELECT * FROM menu_items WHERE category LIKE '%%burger%%' OR category LIKE '%%sandwich%%' OR category LIKE '%%wrap%%';
	- _"Find me coffee available in this cafe."_
	SELECT * FROM menu_items WHERE cm_category = 'beverages' OR category LIKE '%%coffee%%';
	-_"Show me some starters available in this cafe."
	SELECT * FROM menu_items WHERE cm_category = 'starters' OR category LIKE '%%starter%%';

	#### ** Combo Queries (Desserts & Beverages, Under Price Limit)**
	- _"Suggest me some good combos under 600 rupees which include desserts and beverages."_
	SELECT * FROM menu_items 
	WHERE (cm_category IN ('Desserts & Sweets', 'Beverages') OR category LIKE '%%dessert%%' OR category LIKE '%%beverage%%' OR category LIKE '%%mocktail%%' OR category LIKE '%%milkshake%%')
	AND price < 600
	ORDER BY popularity_score DESC;

	#### ** Cuisine-Based Queries (Use cuisines provided above, stricltly)**
	- _"Show me all Italian dishes."_
	SELECT * FROM menu_items WHERE cuisine = 'italian';
	- _"Find me Mediterranean or Greek food."_
	SELECT * FROM menu_items WHERE cuisine = 'mediterranean' OR cuisine = 'greek';

	#### ** Tag-Based Queries (Best Selling, Popular, New Items)**
	- _"Show me only bestseller items."_
	SELECT * FROM menu_items WHERE tag = 'bestseller';
	- _"Show me highly rated items (above 4.5 stars)."_
	SELECT * FROM menu_items WHERE rating > 4.5;
	- _"Find best-rated and most popular dishes."_
	SELECT * FROM menu_items WHERE tag = 'bestrated' OR popularity_score > 4.0;

	#### **Dietary Preferences & Restrictions**
	- _"Show me all vegan options."_
	SELECT * FROM menu_items WHERE dietary_labels LIKE '%%vegan%%';
	- _"I need gluten-free and low-carb dishes."_
	SELECT * FROM menu_items WHERE dietary_labels LIKE '%%gluten-free%%' AND dietary_labels LIKE '%%low-carb%%';
	- _"Find me halal or kosher options."_
	SELECT * FROM menu_items WHERE dietary_labels LIKE '%%halal%%' OR dietary_labels LIKE '%%kosher%%';

	#### ** Spice Level Queries**
	- _"Show me only extra spicy items."_
	SELECT * FROM menu_items WHERE spice_level = 'extra-spicy';
	- _"Find dishes that are either mild or medium spicy."_
	SELECT * FROM menu_items WHERE spice_level IN ('mild', 'medium');

	#### ** Item-Based Queries**
    - _"Show me dishes with chicken."_
   SELECT * FROM menu_items WHERE name LIKE '%%chicken%%' OR name LIKE '%%Chicken%%';
   - _"Find me dishes with bacon."_
   SELECT * FROM menu_items WHERE name LIKE '%%bacon%%' OR name LIKE '%%Bacon%%';


	#### ** Ingredients-Based Queries**
	- _"Exclude items with nuts."_
	SELECT * FROM menu_items WHERE allergens NOT LIKE '%%nuts%%';
	- _"Find me items with truffle or mushroom."_
	SELECT * FROM menu_items WHERE ingredients LIKE '%%truffle%%' OR ingredients LIKE '%%mushroom%%';

	#### **Popularity-Based Queries**
	- _"Show me the most popular dishes."_
	SELECT * FROM menu_items ORDER BY popularity_score DESC LIMIT 10;
	- _"Find top trending items with a rating above 4.5."_
	SELECT * FROM menu_items WHERE rating > 4.5 ORDER BY popularity_score DESC;

	#### ** Complex Queries (Combining Multiple Conditions, while filtering categories strictly use categories provided in the input above)**
	- _"Show me best-selling Italian dishes under 500."_
	SELECT * FROM menu_items WHERE cuisine = 'italian' AND price < 500 AND tag = 'bestseller';
	- _"Find high-rated spicy vegan dishes."_
	SELECT * FROM menu_items WHERE rating > 4.5 AND spice_level = 'spicy' AND dietary_labels LIKE '%%vegan%%';
	- _"Give me gluten-free pastas under 400."_
	SELECT * FROM menu_items WHERE category LIKE '%%pasta%%' AND dietary_labels LIKE '%%gluten-free%%' AND price < 400;
	-_Suggest me some best selling cold coffees in this cafe_
	SELECT * FROM menu_items WHERE category LIKE '%%cold coffee%%' AND category LIKE '%%cold coffee%% ORDER BY DESC popularity_score '

	`, userQuery, categoryList)

	// fmt.Println("Prompt : ", prompt)

	// Call OpenAI API
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens": 500,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to prepare AI request"})
	}

	apiKey := s.Config.OPEN_AI_API_KEY

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create AI request"})
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	currentTime := time.Now()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reach AI service"})
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read AI response"})
	}

	fmt.Println("Time taken to get response from AI : ", time.Since(currentTime))
	// Parse OpenAI response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid AI response format"})
	}

	// Extract AI response
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI response missing choices"})
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI response missing message"})
	}

	content, ok := message["content"].(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI response missing content"})
	}

	// Clean JSON output
	cleanedContent := strings.Trim(content, "`json ")

	// Parse AI response JSON
	var aiResponse map[string]string
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse AI response"})
	}

	// If AI response does not contain SQL, return the response message
	if sqlQuery, exists := aiResponse["sql"]; exists && sqlQuery == "" {
		return c.JSON(fiber.Map{"text": aiResponse["response"], "items": []structures.MenuItem{}})
	}

	fmt.Println("Print SQL query : ", aiResponse["sql"])

	// Execute the generated SQL query
	var menu []structures.MenuItem
	err = s.Db.Raw(aiResponse["sql"]).Scan(&menu).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database query execution failed"})
	}

	// Format response text
	responseText := aiResponse["response"]
	if len(menu) == 0 {
		responseText = "This cafe does not have any matching items."
	}

	// Update in MenuAIRecords table
	menuAIRecord := structures.MenuAIRecords{
		PromptId:     ksuid.New().String(),
		CafeId:       aiRequest.CafeID, // Assuming CafeID is used as UserId here
		UserId:       uint(c.Locals("userId").(float64)),
		Answer:       responseText,
		GeneratedSql: aiResponse["sql"],
		CreatedAt:    time.Now(),
		Prompt:       userQuery,
	}

	if err := s.Db.Create(&menuAIRecord).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save AI record"})
	}

	return c.JSON(fiber.Map{
		"text":  responseText,
		"items": menu,
	})
}
