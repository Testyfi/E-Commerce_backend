package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	models "testify/internal/models"
	httpClient "testify/internal/utility/http"

	"go.mongodb.org/mongo-driver/bson"
)
type statistic struct{

	Date string `json:"date"`
	MaxScore int `json:"maxscore"`
	UserScore int `json:"userscore"`
	UserRank int `json:"userrank"`
	TotalUser int `json:"totaluser"`

}
func UserStatistics(w http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value(models.ContextUser).(models.User)

	if !ok {
		httpClient.RespondError(w, http.StatusBadRequest, "Failed to retrieve user", fmt.Errorf("failed to retrieve user"))
		fmt.Println("some error on user fething statics")
		return 
	}
	allattemptedtest:=UserAttemptedTest(*user.Phone)
	//fmt.Println(allattemptedtest)
	var allstate [] statistic
	for i:=0;i<len(allattemptedtest);i++{
		 var abc statistic
		abc.Date=GetTestDate(allattemptedtest[i])
		abc.MaxScore=TestMaxMarks(allattemptedtest[i])
		abc.TotalUser=TotalStudentAttempted(allattemptedtest[i])
		abc.UserRank=FindUserAndRank(allattemptedtest[i],*user.Phone)
		abc.UserScore=UserMarks(allattemptedtest[i],*user.Phone)
		allstate = append(allstate, abc)
	}
	//UserMaxData()
	httpClient.RespondSuccess(w, allstate)
}
func TestMaxMarks(s string)int  {
	if(s=="Jee Mains"){return 300}
	
    total:=0
	filter := bson.D{{"subject_tags", s}}
	//, {"type", "Single Correct"}
	// Count the number of documents that match the filter
	count, err :=questionCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(count)
	total=int(count)*4
	filter = bson.D{{"subject_tags", s}, {"type", "Single Correct"}}
	count, err =questionCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(count)
	return total-int(count)
}
func UserMaxData(){
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=usermaxscoreCollection.Find(ctx,bson.M{})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println(err)
	}
	type UserMaxScore struct {
	
	
		UserPhone string `json:"userphone"`
		TestName string `json:"testname"`
		TestIndex int `json:"testindex"`
		TotalNumber int `json:"totalnumber"`
		

}
	var usermaxdata []UserMaxScore
if err =cursor.All(ctx,&usermaxdata);err!=nil{
			fmt.Println(err)
		}
		//fmt.Println(usermaxdata)
		

}
func UserMarks(testname string,phone string)int{
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=usermaxscoreCollection.Find(ctx,bson.M{"testname":testname,"userphone":phone})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println(err)
	}
	type UserMaxScore struct {
	
	
		UserPhone string `json:"userphone"`
		TestName string `json:"testname"`
		TestIndex int `json:"testindex"`
		TotalNumber int `json:"totalnumber"`
		

}
	var usermaxdata []UserMaxScore
if err =cursor.All(ctx,&usermaxdata);err!=nil{
			fmt.Println(err)
		}
		//fmt.Println(usermaxdata)
		if(len(usermaxdata)==0){return -1}
		return usermaxdata[0].TotalNumber
}
func UserAttemptedTest(phone string)[] string{

	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=usermaxscoreCollection.Find(ctx,bson.M{"userphone":phone})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println(err)
	}
	type UserMaxScore struct {
	
	
		UserPhone string `json:"userphone"`
		TestName string `json:"testname"`
		TestIndex int `json:"testindex"`
		TotalNumber int `json:"totalnumber"`
		

}
	var usermaxdata []UserMaxScore
if err =cursor.All(ctx,&usermaxdata);err!=nil{
			fmt.Println(err)
		}
		//fmt.Println(usermaxdata)
		var testnames []string
		for i:=0;i<len(usermaxdata);i++{
         testnames=append(testnames,usermaxdata[i].TestName )
		}
		//fmt.Println(testnames)
return testnames
}
func GetTestDate(testname string) string{
	type PaperDetails struct {
		Name       string `json:"Name"`
		Start      string `json:"Start"`
		StartAt    string `json:"StartAt"`
		StartDate  time.Time `json:"StartDate"`
		Difficulty string `json:"Difficulty"`
		Topics     string `json:"Topics"`
		Duration   string `json:"Duration"`
		Prize      string `json:"Prize"`
	}
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=testdetailsCollection.Find(ctx,bson.M{"name":testname})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println(err)
	}
	var papers []PaperDetails
	if err =cursor.All(ctx,&papers);err!=nil{
		fmt.Println(err)
	}
	return papers[0].Start
	

}
func FindUserAndRank(testname string, phone string)int {
    var user models.User
	err := userCollection.FindOne(context.Background(), bson.M{"phone": phone}).Decode(&user)
	if err != nil {
		fmt.Println(err)
		
	}
	return LiveTestRank(testname,user)
}
func TotalStudentAttempted(testname string)int {
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=usermaxscoreCollection.Find(ctx,bson.M{"testname":testname})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println(err)
	}
	type UserMaxScore struct {
	
	
		UserPhone string `json:"userphone"`
		TestName string `json:"testname"`
		TestIndex int `json:"testindex"`
		TotalNumber int `json:"totalnumber"`
		

}
	var usermaxdata []UserMaxScore
if err =cursor.All(ctx,&usermaxdata);err!=nil{
			fmt.Println(err)
		}
		
		return len(usermaxdata)
}


