package repository

import (
	"errors"
	"slices"

	model "github.com/abhinash-kml/go-api-server/internal/models"
)

var (
	ErrSetupFailed     = errors.New("Repository setup failed")
	ErrNoUsers         = errors.New("No user in repository")
	ErrUndefinedUsers  = errors.New("Undefined users")
	ErrZeroLengthSlice = errors.New("Provided slice is of zero length")
	ErrUserDoesntExist = errors.New("Provided users doesn't exist")
)

type UserRepository interface {
	// Initialize
	Setup() error

	// CRUD logics
	GetUsers() ([]model.User, error)
	InsertUsers([]model.User) error
	UpdateUsers([]model.User, []model.User) error
	DeleteUsers([]model.User) error
}

type InMemoryUsersRepository struct {
	users []model.User
}

func NewInMemoryUsersRepository() *InMemoryUsersRepository {
	return &InMemoryUsersRepository{}
}

func (e *InMemoryUsersRepository) Setup() error {
	users := []model.User{
		{Id: 1, Name: "Neo", City: "Kolkata", State: "West Bengal", Country: "India"},
		{Id: 2, Name: "Abhinash", City: "Kolkata", State: "West Bengal", Country: "India"},
		{Id: 3, Name: "Komal", City: "Kolkata", State: "West Bengal", Country: "India"},
		{Id: 4, Name: "Riya", City: "Ranchi", State: "Jharkhand", Country: "India"},
		{Id: 5, Name: "Jyotika", City: "Kolkata", State: "West Bengal", Country: "India"},
		{Id: 6, Name: "Aarav Sharma", City: "Delhi", State: "Delhi", Country: "India"},
		{Id: 7, Name: "Riya Verma", City: "Mumbai", State: "Maharashtra", Country: "India"},
		{Id: 8, Name: "Kunal Mehta", City: "Ahmedabad", State: "Gujarat", Country: "India"},
		{Id: 9, Name: "Sneha Iyer", City: "Chennai", State: "Tamil Nadu", Country: "India"},
		{Id: 10, Name: "Rohit Kulkarni", City: "Pune", State: "Maharashtra", Country: "India"},
		{Id: 11, Name: "Ananya Bose", City: "Kolkata", State: "West Bengal", Country: "India"},
		{Id: 12, Name: "Vikram Singh", City: "Jaipur", State: "Rajasthan", Country: "India"},
		{Id: 13, Name: "Pooja Nair", City: "Kochi", State: "Kerala", Country: "India"},
		{Id: 14, Name: "Arjun Reddy", City: "Hyderabad", State: "Telangana", Country: "India"},
		{Id: 15, Name: "Neha Kapoor", City: "Chandigarh", State: "Chandigarh", Country: "India"},
		{Id: 16, Name: "Siddharth Malhotra", City: "Noida", State: "Uttar Pradesh", Country: "India"},
		{Id: 17, Name: "Mehul Jain", City: "Indore", State: "Madhya Pradesh", Country: "India"},
		{Id: 18, Name: "Tanvi Deshpande", City: "Nagpur", State: "Maharashtra", Country: "India"},
		{Id: 19, Name: "Rahul Chatterjee", City: "Howrah", State: "West Bengal", Country: "India"},
		{Id: 20, Name: "Nikhil Bansal", City: "Gurgaon", State: "Haryana", Country: "India"},
		{Id: 21, Name: "Kavita Joshi", City: "Dehradun", State: "Uttarakhand", Country: "India"},
		{Id: 22, Name: "Amit Mishra", City: "Varanasi", State: "Uttar Pradesh", Country: "India"},
		{Id: 23, Name: "Shreya Ghosh", City: "Siliguri", State: "West Bengal", Country: "India"},
		{Id: 24, Name: "Manish Patel", City: "Surat", State: "Gujarat", Country: "India"},
		{Id: 25, Name: "Ishaan Khanna", City: "Ludhiana", State: "Punjab", Country: "India"},
		{Id: 26, Name: "Divya Agarwal", City: "Aligarh", State: "Uttar Pradesh", Country: "India"},
		{Id: 27, Name: "Saurabh Tiwari", City: "Prayagraj", State: "Uttar Pradesh", Country: "India"},
		{Id: 28, Name: "Ritika Saxena", City: "Bareilly", State: "Uttar Pradesh", Country: "India"},
		{Id: 29, Name: "Karthik Subramanian", City: "Coimbatore", State: "Tamil Nadu", Country: "India"},
		{Id: 30, Name: "Sunita Yadav", City: "Rewari", State: "Haryana", Country: "India"},
		{Id: 31, Name: "Aditya Shetty", City: "Udupi", State: "Karnataka", Country: "India"},
		{Id: 32, Name: "Madhav Rao", City: "Vijayawada", State: "Andhra Pradesh", Country: "India"},
		{Id: 33, Name: "Nandini Kulkarni", City: "Kolhapur", State: "Maharashtra", Country: "India"},
		{Id: 34, Name: "Harshit Arora", City: "Panipat", State: "Haryana", Country: "India"},
		{Id: 35, Name: "Priyanka Thakur", City: "Solan", State: "Himachal Pradesh", Country: "India"},
		{Id: 36, Name: "Akash Dubey", City: "Satna", State: "Madhya Pradesh", Country: "India"},
		{Id: 37, Name: "Meera Pillai", City: "Thiruvananthapuram", State: "Kerala", Country: "India"},
		{Id: 38, Name: "Sanjay Rawat", City: "Haldwani", State: "Uttarakhand", Country: "India"},
		{Id: 39, Name: "Pankaj Soni", City: "Bikaner", State: "Rajasthan", Country: "India"},
		{Id: 40, Name: "Neelam Gupta", City: "Kanpur", State: "Uttar Pradesh", Country: "India"},
		{Id: 41, Name: "Abhishek Roy", City: "Durgapur", State: "West Bengal", Country: "India"},
		{Id: 42, Name: "Rashmi Kulkarni", City: "Nashik", State: "Maharashtra", Country: "India"},
		{Id: 43, Name: "Deepak Yadav", City: "Etawah", State: "Uttar Pradesh", Country: "India"},
		{Id: 44, Name: "Sonal Jain", City: "Ratlam", State: "Madhya Pradesh", Country: "India"},
		{Id: 45, Name: "Varun Khurana", City: "Ambala", State: "Haryana", Country: "India"},
		{Id: 46, Name: "Bhavya Shah", City: "Vadodara", State: "Gujarat", Country: "India"},
		{Id: 47, Name: "Ramesh Iyer", City: "Tiruchirappalli", State: "Tamil Nadu", Country: "India"},
		{Id: 48, Name: "Ankit Saxena", City: "Mathura", State: "Uttar Pradesh", Country: "India"},
		{Id: 49, Name: "Shalini Mehta", City: "Ujjain", State: "Madhya Pradesh", Country: "India"},
		{Id: 50, Name: "Gaurav Malhotra", City: "Patiala", State: "Punjab", Country: "India"},
		{Id: 51, Name: "Ritu Choudhary", City: "Hisar", State: "Haryana", Country: "India"},
		{Id: 52, Name: "Suresh Naidu", City: "Nellore", State: "Andhra Pradesh", Country: "India"},
		{Id: 53, Name: "Kriti Bhat", City: "Mangalore", State: "Karnataka", Country: "India"},
		{Id: 54, Name: "Aman Srivastava", City: "Faizabad", State: "Uttar Pradesh", Country: "India"},
		{Id: 55, Name: "Pallavi Kulkarni", City: "Sangli", State: "Maharashtra", Country: "India"},
		{Id: 56, Name: "Yogesh Pawar", City: "Dhule", State: "Maharashtra", Country: "India"},
		{Id: 57, Name: "Nitin Chauhan", City: "Rohtak", State: "Haryana", Country: "India"},
		{Id: 58, Name: "Sarika Desai", City: "Valsad", State: "Gujarat", Country: "India"},
		{Id: 59, Name: "Alok Pandey", City: "Ballia", State: "Uttar Pradesh", Country: "India"},
		{Id: 60, Name: "Monika Sen", City: "Raipur", State: "Chhattisgarh", Country: "India"},
		{Id: 61, Name: "Ravindra Patil", City: "Jalgaon", State: "Maharashtra", Country: "India"},
		{Id: 62, Name: "Kishore R", City: "Salem", State: "Tamil Nadu", Country: "India"},
		{Id: 63, Name: "Anjali Thapa", City: "Gangtok", State: "Sikkim", Country: "India"},
		{Id: 64, Name: "Mohit Sethi", City: "Jammu", State: "Jammu and Kashmir", Country: "India"},
		{Id: 65, Name: "Rekha Devi", City: "Gaya", State: "Bihar", Country: "India"},
		{Id: 66, Name: "Shubham Tandon", City: "Meerut", State: "Uttar Pradesh", Country: "India"},
		{Id: 67, Name: "Vinod Kori", City: "Katni", State: "Madhya Pradesh", Country: "India"},
		{Id: 68, Name: "Pritam Das", City: "Agartala", State: "Tripura", Country: "India"},
		{Id: 69, Name: "Lalit Joshi", City: "Pithoragarh", State: "Uttarakhand", Country: "India"},
		{Id: 70, Name: "Kiran Patnaik", City: "Cuttack", State: "Odisha", Country: "India"},
		{Id: 71, Name: "Seema Rani", City: "Muzaffarpur", State: "Bihar", Country: "India"},
		{Id: 72, Name: "Arvind Chawla", City: "Sirsa", State: "Haryana", Country: "India"},
		{Id: 73, Name: "Snehal Gawade", City: "Satara", State: "Maharashtra", Country: "India"},
		{Id: 74, Name: "Naresh Kumar", City: "Sikar", State: "Rajasthan", Country: "India"},
		{Id: 75, Name: "Bharat Reddy", City: "Anantapur", State: "Andhra Pradesh", Country: "India"},
		{Id: 76, Name: "Payal Arora", City: "Sonipat", State: "Haryana", Country: "India"},
		{Id: 77, Name: "Siddique Khan", City: "Bhopal", State: "Madhya Pradesh", Country: "India"},
		{Id: 78, Name: "Ila Sengupta", City: "Bardhaman", State: "West Bengal", Country: "India"},
		{Id: 79, Name: "Naveen Rao", City: "Shimoga", State: "Karnataka", Country: "India"},
		{Id: 80, Name: "Rupal Doshi", City: "Palanpur", State: "Gujarat", Country: "India"},
		{Id: 81, Name: "Hemant Pathak", City: "Chitrakoot", State: "Uttar Pradesh", Country: "India"},
		{Id: 82, Name: "Ayesha Rahman", City: "Malda", State: "West Bengal", Country: "India"},
		{Id: 83, Name: "Pradeep N", City: "Hosur", State: "Tamil Nadu", Country: "India"},
		{Id: 84, Name: "Suman Deb", City: "Silchar", State: "Assam", Country: "India"},
		{Id: 85, Name: "Rohini Kulkarni", City: "Akola", State: "Maharashtra", Country: "India"},
		{Id: 86, Name: "Ajay Parmar", City: "Godhra", State: "Gujarat", Country: "India"},
		{Id: 87, Name: "Ranjit Kaur", City: "Moga", State: "Punjab", Country: "India"},
		{Id: 88, Name: "Nilesh Raut", City: "Wardha", State: "Maharashtra", Country: "India"},
		{Id: 89, Name: "Farhan Ali", City: "Moradabad", State: "Uttar Pradesh", Country: "India"},
		{Id: 90, Name: "Kavya Menon", City: "Thrissur", State: "Kerala", Country: "India"},
		{Id: 91, Name: "Rajesh Yadav", City: "Azamgarh", State: "Uttar Pradesh", Country: "India"},
		{Id: 92, Name: "Tulsi Devi", City: "Madhubani", State: "Bihar", Country: "India"},
		{Id: 93, Name: "Girish Hegde", City: "Sirsi", State: "Karnataka", Country: "India"},
		{Id: 94, Name: "Satyam Tripathi", City: "Rewa", State: "Madhya Pradesh", Country: "India"},
		{Id: 95, Name: "Neetu Sharma", City: "Palwal", State: "Haryana", Country: "India"},
		{Id: 96, Name: "Jitendra Solanki", City: "Neemuch", State: "Madhya Pradesh", Country: "India"},
		{Id: 97, Name: "Abdul Latif", City: "Darbhanga", State: "Bihar", Country: "India"},
		{Id: 98, Name: "Komal Suri", City: "Una", State: "Himachal Pradesh", Country: "India"},
		{Id: 99, Name: "Suresh Goud", City: "Mahbubnagar", State: "Telangana", Country: "India"},
		{Id: 100, Name: "Aditi Kulshreshtha", City: "Gwalior", State: "Madhya Pradesh", Country: "India"},
		{Id: 101, Name: "Ranjan Das", City: "Kokrajhar", State: "Assam", Country: "India"},
		{Id: 102, Name: "Keshav Joshi", City: "Bundi", State: "Rajasthan", Country: "India"},
		{Id: 103, Name: "Preeti Mahajan", City: "Pathankot", State: "Punjab", Country: "India"},
		{Id: 104, Name: "Nagaraj K", City: "Chitradurga", State: "Karnataka", Country: "India"},
		{Id: 105, Name: "Ramesh Khatri", City: "Alwar", State: "Rajasthan", Country: "India"},
	}

	e.users = users

	return nil
}

func (e *InMemoryUsersRepository) GetUsers() ([]model.User, error) {
	if len(e.users) <= 0 {
		return nil, ErrNoUsers
	}

	return e.users, nil
}

func (e *InMemoryUsersRepository) InsertUsers(users []model.User) error {
	if len(users) <= 0 {
		return ErrZeroLengthSlice
	}

	for _, value := range users {
		e.users = append(e.users, value)
	}

	return nil
}

func (e *InMemoryUsersRepository) UpdateUsers(old, new []model.User) error {
	if len(old) <= 0 || len(new) <= 0 {
		return ErrZeroLengthSlice
	}

	for _, value := range e.users {
		for i := 0; i < len(old); i++ {
			if value.Id == old[i].Id {
				value = new[i]
			}
		}
	}

	return nil
}

func (e *InMemoryUsersRepository) DeleteUsers(users []model.User) error {
	if len(users) <= 0 {
		return ErrZeroLengthSlice
	}

	e.users = slices.DeleteFunc(e.users, func(u model.User) bool {
		for _, value := range users {
			if u.Id == value.Id {
				return true
			}
		}

		return false
	})

	return nil
}
