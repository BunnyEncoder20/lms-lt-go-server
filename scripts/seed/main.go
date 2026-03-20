package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"go-server/internal/database/db"
	"go-server/internal/models"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// Helper to hash passwords quickly during seed
func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes)
}

func main() {
	log.Println("🌱 Seeding database...")

	// Connecting to the database
	if err := godotenv.Load("./.env"); err != nil {
		log.Println("❗️ WARN: could not load .env file, DB_URL is likely empty", err)
	}
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "local_lms.db" // Fallback
	}
	conn, err := sql.Open("sqlite", dbUrl)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	ctx := context.Background()

	// Default values for required fields in users table
	defaultUserParams := db.UpsertUserParams{
		Title:        "Mr/Ms",
		Gender:       "M",
		Band:         "B1",
		Grade:        "G1",
		Ic:           "IC1",
		Sbg:          "SBG1",
		Bu:           "BU1",
		Segment:      "S1",
		Department:   "D1",
		BaseLocation: "Location1",
	}

	// ── Users ─────────────────────────────────────────────────────────────────
	// Fixed IDs for consistent references
	adminID := uuid.MustParse("e83e0a1b-a580-41d6-96ef-79bc3b476781")
	manager1ID := uuid.MustParse("017ee8c7-9424-4a38-bca0-34add327c28a")
	manager2ID := uuid.MustParse("e0959662-e6d1-4c0d-9676-59c119e47fda")
	emp1ID := uuid.MustParse("a775afc2-8c9e-4eda-9c5e-cb3e1220d7f1")
	emp2ID := uuid.MustParse("d47773c4-12d1-4570-99a4-50f1bd11fb7a")
	emp3ID := uuid.MustParse("062d05b9-475b-4ba2-8a33-452c663747b1")
	cd1ID := uuid.MustParse("3ee57473-98de-4d8e-a052-3c375d67053a")

	upsertUser := func(params db.UpsertUserParams) db.User {
		// Apply defaults
		if params.Title == "" {
			params.Title = defaultUserParams.Title
		}
		if params.Gender == "" {
			params.Gender = defaultUserParams.Gender
		}
		if params.Band == "" {
			params.Band = defaultUserParams.Band
		}
		if params.Grade == "" {
			params.Grade = defaultUserParams.Grade
		}
		if params.Ic == "" {
			params.Ic = defaultUserParams.Ic
		}
		if params.Sbg == "" {
			params.Sbg = defaultUserParams.Sbg
		}
		if params.Bu == "" {
			params.Bu = defaultUserParams.Bu
		}
		if params.Segment == "" {
			params.Segment = defaultUserParams.Segment
		}
		if params.Department == "" {
			params.Department = defaultUserParams.Department
		}
		if params.BaseLocation == "" {
			params.BaseLocation = defaultUserParams.BaseLocation
		}
		user, err := queries.UpsertUser(ctx, params)
		if err != nil {
			log.Fatalf("Failed to upsert user %s: %v", params.PesNumber, err)
		}
		log.Printf("  ✅ User: %s — %s %s", user.PesNumber, user.FirstName, user.LastName)
		return user
	}

	admin := upsertUser(db.UpsertUserParams{
		ID:        adminID,
		PesNumber: "PES-001",
		Password:  hashPassword("Admin123"),
		FirstName: "System",
		LastName:  "Administrator",
		Email:     "admin@skillsync.com",
		Role:      models.RoleAdmin,
		Cluster:   sql.NullString{String: "IT", Valid: true},
	})

	manager1 := upsertUser(db.UpsertUserParams{
		ID:        manager1ID,
		PesNumber: "PES-002",
		Password:  hashPassword("Manager123"),
		FirstName: "James",
		LastName:  "Wilson",
		Email:     "manager@skillsync.com",
		Role:      models.RoleManager,
		Cluster:   sql.NullString{String: "Operations", Valid: true},
	})

	manager2 := upsertUser(db.UpsertUserParams{
		ID:        manager2ID,
		PesNumber: "PES-005",
		Password:  hashPassword("Manager123"),
		FirstName: "Emily",
		LastName:  "Davis",
		Email:     "emily@skillsync.com",
		Role:      models.RoleManager,
		Cluster:   sql.NullString{String: "HR", Valid: true},
	})

	emp1 := upsertUser(db.UpsertUserParams{
		ID:        emp1ID,
		PesNumber: "PES-003",
		Password:  hashPassword("Employee123"),
		FirstName: "Sarah",
		LastName:  "Johnson",
		Email:     "sarah@skillsync.com",
		Role:      models.RoleEmployee,
		Cluster:   sql.NullString{String: "Operations", Valid: true},
		IsID:      uuid.NullUUID{UUID: manager1.ID, Valid: true},
	})

	emp2 := upsertUser(db.UpsertUserParams{
		ID:        emp2ID,
		PesNumber: "PES-004",
		Password:  hashPassword("Employee123"),
		FirstName: "Michael",
		LastName:  "Chen",
		Email:     "michael@skillsync.com",
		Role:      models.RoleEmployee,
		Cluster:   sql.NullString{String: "Operations", Valid: true},
		IsID:      uuid.NullUUID{UUID: manager1.ID, Valid: true},
	})

	emp3 := upsertUser(db.UpsertUserParams{
		ID:        emp3ID,
		PesNumber: "PES-006",
		Password:  hashPassword("Employee123"),
		FirstName: "Priya",
		LastName:  "Sharma",
		Email:     "priya@skillsync.com",
		Role:      models.RoleEmployee,
		Cluster:   sql.NullString{String: "HR", Valid: true},
		IsID:      uuid.NullUUID{UUID: manager2.ID, Valid: true},
	})

	cd1 := upsertUser(db.UpsertUserParams{
		ID:        cd1ID,
		PesNumber: "PES-007",
		Password:  hashPassword("CD123"),
		FirstName: "Sarah",
		LastName:  "Blake",
		Email:     "sarah.blake@skillsync.com",
		Role:      models.RoleCourseDirector,
		Cluster:   sql.NullString{String: "L&D", Valid: true},
	})

	// ── Trainings ─────────────────────────────────────────────────────────────
	now := time.Now()
	future := func(days int) time.Time { return now.AddDate(0, 0, days) }
	past := func(days int) time.Time { return now.AddDate(0, 0, -days) }

	// Default values for trainings
	defaultTrainingParams := db.UpsertTrainingParams{
		HrProgramID:    uuid.New(),
		ModeOfDelivery: models.InPerson,
		InstructorName: "System",
		CreatedByID:    admin.ID,
	}

	trainingsData := []db.UpsertTrainingParams{
		{
			ID:           uuid.New(),
			Title:        "Leadership Excellence Program",
			Description:  sql.NullString{String: "Develop leadership and team management skills for managerial roles.", Valid: true},
			Category:     models.TrainingBehavioral,
			StartDate:    future(10),
			EndDate:      future(11),
			Location:     sql.NullString{String: "Conference Room A", Valid: true},
			PreReadUri:   sql.NullString{String: "https://example.com/leadership-prework.pdf", Valid: true},
			DeadlineDays: 2,
		},
		{
			ID:           uuid.New(),
			Title:        "Advanced Excel & Data Analysis",
			Description:  sql.NullString{String: "Master pivot tables, VLOOKUP, and data visualization in Excel.", Valid: true},
			Category:     models.TrainingTechnical,
			StartDate:    future(15),
			EndDate:      future(15),
			Location:     sql.NullString{String: "Training Lab 2", Valid: true},
			PreReadUri:   sql.NullString{String: "https://example.com/excel-guide.pdf", Valid: true},
			DeadlineDays: 1,
		},
		{
			ID:           uuid.New(),
			Title:        "Effective Communication & Presentation",
			Description:  sql.NullString{String: "Build confident public speaking and stakeholder communication skills.", Valid: true},
			Category:     models.TrainingBehavioral,
			StartDate:    future(20),
			EndDate:      future(21),
			VirtualLink:  sql.NullString{String: "https://teams.microsoft.com/skillsync-comm", Valid: true},
			DeadlineDays: 2,
		},
		{
			ID:           uuid.New(),
			Title:        "Cybersecurity Awareness",
			Description:  sql.NullString{String: "Mandatory security awareness training covering phishing and data protection.", Valid: true},
			Category:     models.TrainingTechnical,
			StartDate:    future(5),
			EndDate:      future(5),
			Location:     sql.NullString{String: "Auditorium", Valid: true},
			PreReadUri:   sql.NullString{String: "https://example.com/cybersecurity-intro.pdf", Valid: true},
			DeadlineDays: 1,
		},
		{
			ID:           uuid.New(),
			Title:        "Project Management Fundamentals",
			Description:  sql.NullString{String: "Introduction to Agile, Scrum, and PMBOK frameworks.", Valid: true},
			Category:     models.TrainingTechnical,
			StartDate:    past(30),
			EndDate:      past(29),
			Location:     sql.NullString{String: "Conference Room B", Valid: true},
			PreReadUri:   sql.NullString{String: "https://example.com/pm-basics.pdf", Valid: true},
			DeadlineDays: 3,
		},
		{
			ID:           uuid.New(),
			Title:        "Conflict Resolution & Negotiation",
			Description:  sql.NullString{String: "Techniques for resolving workplace conflicts constructively.", Valid: true},
			Category:     models.TrainingBehavioral,
			StartDate:    past(60),
			EndDate:      past(59),
			VirtualLink:  sql.NullString{String: "https://teams.microsoft.com/skillsync-conflict", Valid: true},
			DeadlineDays: 2,
		},
	}

	createdTrainings := make(map[string]db.Training)
	for _, t := range trainingsData {
		// Set defaults
		t.HrProgramID = defaultTrainingParams.HrProgramID
		t.ModeOfDelivery = defaultTrainingParams.ModeOfDelivery
		t.InstructorName = defaultTrainingParams.InstructorName
		t.CreatedByID = defaultTrainingParams.CreatedByID
		t.MappedCategory = string(t.Category)

		training, err := queries.UpsertTraining(ctx, t)
		if err != nil {
			log.Fatalf("Failed to upsert training %s: %v", t.Title, err)
		}
		createdTrainings[training.Title] = training
		log.Printf("  ✅ Training: %s", training.Title)
	}

	// ── Nominations ───────────────────────────────────────────────────────────
	nominations := []db.UpsertNominationParams{
		{
			ID:            uuid.New(),
			UserID:        emp1.ID,
			TrainingID:    createdTrainings["Advanced Excel & Data Analysis"].ID,
			NominatedByID: emp1.ID,
			Status:        models.NomPending,
		},
		{
			ID:            uuid.New(),
			UserID:        emp1.ID,
			TrainingID:    createdTrainings["Leadership Excellence Program"].ID,
			NominatedByID: manager1.ID,
			Status:        models.NomApproved,
		},
		{
			ID:            uuid.New(),
			UserID:        emp2.ID,
			TrainingID:    createdTrainings["Cybersecurity Awareness"].ID,
			NominatedByID: manager1.ID,
			Status:        models.NomApproved,
		},
		{
			ID:            uuid.New(),
			UserID:        emp2.ID,
			TrainingID:    createdTrainings["Project Management Fundamentals"].ID,
			NominatedByID: manager1.ID,
			Status:        models.NomCompleted,
		},
		{
			ID:            uuid.New(),
			UserID:        emp3.ID,
			TrainingID:    createdTrainings["Conflict Resolution & Negotiation"].ID,
			NominatedByID: manager2.ID,
			Status:        models.NomAttended,
		},
		{
			ID:            uuid.New(),
			UserID:        emp3.ID,
			TrainingID:    createdTrainings["Effective Communication & Presentation"].ID,
			NominatedByID: emp3.ID,
			Status:        models.NomPending,
		},
	}

	for _, n := range nominations {
		_, err := queries.UpsertNomination(ctx, n)
		if err != nil {
			log.Fatalf("Failed to upsert nomination: %v", err)
		}
	}
	log.Printf("  ✅ Seeded %d nominations", len(nominations))

	// ── Courses ─────────────────────────────────────────────────────────────
	outcomes1, _ := json.Marshal([]string{
		"Write Python scripts for data manipulation",
		"Use Pandas DataFrames for analysis",
		"Create visualizations with Matplotlib",
		"Automate repetitive data tasks",
	})
	course1, err := queries.UpsertCourse(ctx, db.UpsertCourseParams{
		ID:                 uuid.New(),
		Title:              "Python for Data Analysis",
		Description:        sql.NullString{String: "A comprehensive course covering Python fundamentals, data manipulation with Pandas, and data visualization with Matplotlib for business analysts.", Valid: true},
		AuthorID:           uuid.NullUUID{UUID: cd1.ID, Valid: true},
		Status:             models.CoursePublished,
		Category:           models.TrainingTechnical,
		EstimatedDurations: sql.NullInt64{Int64: 120, Valid: true},
		LearningOutcomes:   string(outcomes1),
		IsStrictSequencing: true,
		Version:            1,
		PublishedAt:        sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		log.Fatalf("Failed to upsert course 1: %v", err)
	}
	log.Printf("  ✅ Course: %s", course1.Title)

	// Modules for Course 1
	mod1, _ := queries.UpsertCourseModule(ctx, db.UpsertCourseModuleParams{
		ID:            uuid.New(),
		Title:         "Getting Started",
		CourseID:      course1.ID,
		Description:   sql.NullString{String: "Introduction and environment setup", Valid: true},
		SequenceOrder: 1,
	})
	queries.UpsertLesson(ctx, db.UpsertLessonParams{
		ID:              uuid.New(),
		Title:           "What is Python?",
		ModuleID:        mod1.ID,
		ContentType:     models.LessonRichText,
		RichTextContent: sql.NullString{String: "<h2>What is Python?</h2><p>Python is a versatile, high-level programming language known for its readability and broad ecosystem of libraries. It is widely used in data science, web development, automation, and intelligence.</p><h3>Why Python for Data?</h3><ul><li>Simple syntax that reads like English</li><li>Massive library ecosystem (Pandas, NumPy, Matplotlib)</li><li>Strong community support</li><li>Used by Google, Netflix, NASA, and more</li></ul>", Valid: true},
		SequenceOrder:   1,
	})
	queries.UpsertLesson(ctx, db.UpsertLessonParams{
		ID:              uuid.New(),
		Title:           "Setup & Installation Guide",
		ModuleID:        mod1.ID,
		ContentType:     models.LessonRichText,
		RichTextContent: sql.NullString{String: "<h2>Setting Up Your Environment</h2><p>Follow these steps to install Python and set up your development environment:</p><ol><li>Download Python 3.12 from <strong>python.org</strong></li><li>Install VS Code as your editor</li><li>Create a virtual environment: <code>python -m venv myenv</code></li><li>Install packages: <code>pip install pandas matplotlib jupyter</code></li></ol><h3>Verify Installation</h3><pre><code>python --version\npip list</code></pre>", Valid: true},
		SequenceOrder:   2,
		DurationMinutes: sql.NullInt64{Int64: 5, Valid: true},
	})

	mod2, _ := queries.UpsertCourseModule(ctx, db.UpsertCourseModuleParams{
		ID:            uuid.New(),
		Title:         "Core Python Concepts",
		CourseID:      course1.ID,
		Description:   sql.NullString{String: "Variables, functions, and data structures", Valid: true},
		SequenceOrder: 2,
	})
	queries.UpsertLesson(ctx, db.UpsertLessonParams{
		ID:              uuid.New(),
		Title:           "Variables & Data Types",
		ModuleID:        mod2.ID,
		ContentType:     models.LessonRichText,
		RichTextContent: sql.NullString{String: "<h2>Variables & Data Types</h2><p>Python supports several built-in data types:</p><ul><li><strong>int</strong> — Integers (42)</li><li><strong>float</strong> — Decimals (3.14)</li><li><strong>str</strong> — Strings (\"hello\")</li><li><strong>bool</strong> — Boolean (True/False)</li><li><strong>list</strong> — Ordered collection [1, 2, 3]</li><li><strong>dict</strong> — Key-value pairs {\"name\": \"Alice\"}</li></ul><h3>Type Conversion</h3><pre><code>age = int(\"25\")\nprice = float(\"19.99\")</code></pre>", Valid: true},
		SequenceOrder:   1,
	})
	queries.UpsertLesson(ctx, db.UpsertLessonParams{
		ID:              uuid.New(),
		Title:           "Functions Deep Dive",
		ModuleID:        mod2.ID,
		ContentType:     models.LessonRichText,
		RichTextContent: sql.NullString{String: "<h2>Functions in Python</h2><p>Functions are reusable blocks of code that perform a specific task.</p><pre><code>def greet(name, greeting=\"Hello\"):\n    return f\"{greeting}, {name}!\"\n\nresult = greet(\"Alice\")\nprint(result)  # Hello, Alice!</code></pre><h3>Key Concepts</h3><ul><li>Default parameters</li><li>*args and **kwargs</li><li>Return values</li><li>Lambda functions</li></ul>", Valid: true},
		SequenceOrder:   2,
		DurationMinutes: sql.NullInt64{Int64: 8, Valid: true},
	})
	queries.UpsertLesson(ctx, db.UpsertLessonParams{
		ID:              uuid.New(),
		Title:           "Python Cheatsheet",
		ModuleID:        mod2.ID,
		ContentType:     models.LessonRichText,
		RichTextContent: sql.NullString{String: "<h2>Python Quick Reference</h2><table><tr><th>Operation</th><th>Syntax</th></tr><tr><td>Print</td><td><code>print(\"hello\")</code></td></tr><tr><td>List comprehension</td><td><code>[x**2 for x in range(10)]</code></td></tr><tr><td>Dictionary</td><td><code>{\"key\": \"value\"}</code></td></tr><tr><td>F-string</td><td><code>f\"Name: {name}\"</code></td></tr><tr><td>Try/Except</td><td><code>try: ... except: ...</code></td></tr></table>", Valid: true},
		SequenceOrder:   3,
	})

	outcomes2, _ := json.Marshal([]string{
		"Deliver confident presentations",
		"Write clear and concise emails",
		"Handle difficult conversations professionally",
	})
	course2, err := queries.UpsertCourse(ctx, db.UpsertCourseParams{
		ID:                 uuid.New(),
		Title:              "Advanced Communication Skills",
		Description:        sql.NullString{String: "Master the art of professional communication, from boardroom presentations to cross-functional collaboration.", Valid: true},
		AuthorID:           uuid.NullUUID{UUID: cd1.ID, Valid: true},
		Status:             models.CourseDraft,
		Category:           models.TrainingBehavioral,
		EstimatedDurations: sql.NullInt64{Int64: 90, Valid: true},
		LearningOutcomes:   string(outcomes2),
		IsStrictSequencing: false,
		Version:            1,
	})
	if err != nil {
		log.Fatalf("Failed to upsert course 2: %v", err)
	}
	log.Printf("  ✅ Course: %s", course2.Title)

	mod3, _ := queries.UpsertCourseModule(ctx, db.UpsertCourseModuleParams{
		ID:            uuid.New(),
		Title:         "Speaking Fundamentals",
		CourseID:      course2.ID,
		Description:   sql.NullString{String: "Build a strong foundation for public speaking", Valid: true},
		SequenceOrder: 1,
	})
	queries.UpsertLesson(ctx, db.UpsertLessonParams{
		ID:              uuid.New(),
		Title:           "Breathing & Posture Techniques",
		ModuleID:        mod3.ID,
		ContentType:     models.LessonRichText,
		RichTextContent: sql.NullString{String: "<h2>Breathing & Posture</h2><p>Great speakers control their breath. Try the 4-7-8 technique:</p><ol><li>Inhale for 4 seconds</li><li>Hold for 7 seconds</li><li>Exhale for 8 seconds</li></ol><h3>Posture Tips</h3><ul><li>Stand with feet shoulder-width apart</li><li>Keep shoulders back and relaxed</li><li>Maintain eye contact with the audience</li></ul>", Valid: true},
		SequenceOrder:   1,
		DurationMinutes: sql.NullInt64{Int64: 5, Valid: true},
	})
	queries.UpsertLesson(ctx, db.UpsertLessonParams{
		ID:              uuid.New(),
		Title:           "Structuring Your Message",
		ModuleID:        mod3.ID,
		ContentType:     models.LessonRichText,
		RichTextContent: sql.NullString{String: "<h2>Message Structure</h2><p>Use the PREP framework for any presentation:</p><ul><li><strong>P</strong>oint — State your main argument</li><li><strong>R</strong>eason — Explain why</li><li><strong>E</strong>xample — Give evidence</li><li><strong>P</strong>oint — Restate the conclusion</li></ul>", Valid: true},
		SequenceOrder:   2,
		DurationMinutes: sql.NullInt64{Int64: 10, Valid: true},
	})

	// ── Course Assignments ────────────────────────────────────────────────────
	queries.UpsertCourseAssignment(ctx, db.UpsertCourseAssignmentParams{
		ID:            uuid.New(),
		CourseID:      course1.ID,
		UserID:        uuid.NullUUID{UUID: emp1.ID, Valid: true},
		AssignedByID:  uuid.NullUUID{UUID: admin.ID, Valid: true},
		Status:        models.AssignmentNotStarted,
		CourseVersion: 1,
	})

	assign2, _ := queries.UpsertCourseAssignment(ctx, db.UpsertCourseAssignmentParams{
		ID:                 uuid.New(),
		CourseID:           course1.ID,
		UserID:             uuid.NullUUID{UUID: emp2.ID, Valid: true},
		AssignedByID:       uuid.NullUUID{UUID: admin.ID, Valid: true},
		Status:             models.AssignmentInProgress,
		ProgressPercentage: 60.0,
		CourseVersion:      1,
	})

	// Mark first 3 lessons as completed for emp2
	lessons, _ := queries.ListLessonsByCourse(ctx, course1.ID)
	for i := 0; i < len(lessons) && i < 3; i++ {
		queries.UpsertLessonProgress(ctx, db.UpsertLessonProgressParams{
			ID:           uuid.New(),
			AssignmentID: assign2.ID,
			LessonID:     lessons[i].ID,
			IsCompleted:  true,
			CompletedAt:  sql.NullTime{Time: now, Valid: true},
		})
	}

	log.Println("\n📋 Default Credentials:")
	log.Println("  Admin    → PES-001 / Admin123")
	log.Println("  Manager  → PES-002 / Manager123")
	log.Println("  Manager  → PES-005 / Manager123")
	log.Println("  Employee → PES-003 / Employee123")
	log.Println("  Employee → PES-004 / Employee123")
	log.Println("  Employee → PES-006 / Employee123")
	log.Println("  CD       → PES-007 / CD123")
	log.Println("\n🌱 Seeding complete!")
}
