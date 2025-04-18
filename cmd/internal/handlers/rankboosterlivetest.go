package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	models "testify/internal/models"
	httpClient "testify/internal/utility/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetLiveTestQuestion(w http.ResponseWriter, r *http.Request) {
	//httpClient.RespondSuccess(w, "success")
	//fmt.Println("success")
	var t struct {
	Testname string `json:"testname"`
	}
    
	err := json.NewDecoder(r.Body).Decode(&t)
    //fmt.Println(t.Testname)
	if err != nil {
		fmt.Println("error "+err.Error())
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid testname", err)
		return
	}
	//
	//fmt.Println(LiveTestTimeValidation(t.Testname))
	if(LiveTestTimeValidation(t.Testname)){
        
		var index=LiveTestFindPaperIndex(t.Testname);
        //var index=10
		ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
		cursor,err:=questionCollection.Find(ctx,bson.M{"subject_tags":t.Testname})
		defer cursor.Close(ctx)
		if err !=nil{
	
			httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
			return
		}
		type QuestionType struct {
	
	
		ID             string        `json:"_id"`
		Correctanswer  string        `json:"correctanswer"`
		Correctanswers []string      `json:"correctanswers"`
		CreatedAt      time.Time     `json:"created_at"`
		Images         []interface{} `json:"images"`
		List1          interface{}   `json:"list1"`
		List2          interface{}   `json:"list2"`
		Options        []struct {
			Image string `json:"image"`
			Text  string `json:"text"`
		} `json:"options"`
		QID         string      `json:"q_id"`
		Question    string      `json:"question"`
		Solution    string      `json:"solution"`
		SubjectTags []string    `json:"subject_tags"`
		Type        string      `json:"type"`
		Usedby      interface{} `json:"usedby"`
	
}
 var questions []QuestionType
if err =cursor.All(ctx,&questions);err!=nil{
			httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
			return
		}

/*
		var questions []bson.M
		if err =cursor.All(ctx,&questions);err!=nil{
			httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
			return
		}
	*/	
		 //fmt.Println(index)
		questions[index].Correctanswer="";
		var str[]string
		questions[index].Correctanswers=str
	
		httpClient.RespondSuccess(w, questions[index])
		return
	}
	indianTimeZone, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Error loading Indian time zone:", err)
		return
	}

	// Get the current time in the Indian time zone
	indianTime := time.Now().In(indianTimeZone)
	httpClient.RespondSuccess(w, indianTime)
	return 
   //fmt.Println("At End")
}
func LiveTestResponse(w http.ResponseWriter, r *http.Request){

	user, ok := r.Context().Value(models.ContextUser).(models.User)

	if !ok {
		httpClient.RespondError(w, http.StatusBadRequest, "Failed to retrieve user", fmt.Errorf("failed to retrieve user"))
		return
	}
	var t struct {
		TestAnswer string `json:"testanswer"`
		TestName string `json:"testname"`
		}
	
		err := json.NewDecoder(r.Body).Decode(&t)
	
		if err != nil {
			httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid testname", err)
			return
		}
		if(LiveTestTimeValidation(t.TestName)){
		var index=LiveTestFindPaperIndex(t.TestName)-1;
		if(index>=0){
			var userresponce struct{

				UserData models.User `json:"user"`
				TestName string `json:"testname"`
				TestIndex int `json:"testindex"`
				Answer string `json:"answer"`
				Value bool `json:"value"`
			}
			
			userresponce.UserData=user;
		
			userresponce.Answer=t.TestAnswer;
			userresponce.TestIndex=index;
			
			userresponce.TestName=t.TestName
			userresponce.Value=MatchAnswer(t.TestAnswer,ReturnAnswer(t.TestName,index))
			fmt.Println(GetNumberFromResponse(index,t.TestAnswer,t.TestName))
			testpaperCollection.InsertOne(context.TODO(), userresponce)
			
			
			filter := bson.M{"userphone": user.Phone,"testindex":-1,"testname":t.TestName}
			update := bson.M{"$inc": bson.M{"totalnumber": GetNumberFromResponse(index,t.TestAnswer,t.TestName)}}
			opts := options.Update().SetUpsert(true)

	// Update the document with the specified filter and update
	usermaxscoreCollection.UpdateOne(context.Background(), filter, update, opts)
	        
	//filter := bson.D{{"username", "john_doe"}}

	// Find documents that match the filter
	
   
	// Iterate over the cursor to process the results
	
	
			 var res struct {

				TotalUser int64 `json:"allstudent"`
				Rank int `json:"rank"`
			 }
			 res.Rank=LiveTestRank(t.TestName,user)
			 filter = bson.M{"testindex":-1,"testname":t.TestName}
			 

	// Get the size of the collection with the specified filter
	counter ,err:= usermaxscoreCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
			 res.TotalUser=counter
				httpClient.RespondSuccess(w,res)
			
            
           
	
		}
	}

		

}
func UserRank(w http.ResponseWriter, r *http.Request){
	user, ok := r.Context().Value(models.ContextUser).(models.User)

	if !ok {
		httpClient.RespondError(w, http.StatusBadRequest, "Failed to retrieve user", fmt.Errorf("failed to retrieve user"))
		return
	}
	var t struct {
		
		TestName string `json:"testname"`
		}
	

		err := json.NewDecoder(r.Body).Decode(&t)
	
		if err != nil {
			httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid testname", err)
			return
		}
		var res struct {

			TestRank int `json:"testrank"`
			TestMarks int `json:"testmarks"`
		}
		res.TestRank=LiveTestRank(t.TestName,user)
		res.TestMarks=FindUserTotalNumber(t.TestName,user)
		httpClient.RespondSuccess(w,res)

}
func TotalUsers(w http.ResponseWriter, r *http.Request){

	var t struct {
		
		TestName string `json:"testname"`
		}
	

		err := json.NewDecoder(r.Body).Decode(&t)
		ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=totaluserCollection.Find(ctx,bson.M{"testname":t.TestName})
	defer cursor.Close(ctx)
	if err !=nil{

		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}
	type UserMaxScore struct {
	
	
		
		TestName string `json:"testname"`
		
		TotalUser int `json:"totaluser"`
		

}
	var usermaxdata []UserMaxScore
if err =cursor.All(ctx,&usermaxdata);err!=nil{
			httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
			return
		}
		httpClient.RespondSuccess(w,usermaxdata[0].TotalUser)

	
}
func IncrementUser(w http.ResponseWriter, r *http.Request){

	var t struct {
		
		TestName string `json:"testname"`
		}
	
        
		err := json.NewDecoder(r.Body).Decode(&t)
		if err!=nil {fmt.Println(err)}
		    filter := bson.M{"testname": t.TestName}
			update := bson.M{"$inc": bson.M{"totaluser": 1}}
			opts := options.Update().SetUpsert(true)

	// Update the document with the specified filter and update
	totaluserCollection.UpdateOne(context.Background(), filter, update, opts)
	
	httpClient.RespondSuccess(w,"success")
}
func DeleteLiveTestAllUserData(w http.ResponseWriter, r *http.Request){

	
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=usermaxscoreCollection.Find(ctx,bson.M{})
	defer cursor.Close(ctx)
	if err !=nil{

		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}
	type UserMaxScore struct {
	
	
		UserPhone string `json:"userphone"`
		TestName string `json:"testname"`
		TestIndex int `json:"testindex"`
		TotalNumber int `json:"totalnumber"`
		

}
	var usermaxdata []UserMaxScore
if err =cursor.All(ctx,&usermaxdata);err!=nil{
			httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
			return
		}
	// Create an empty filter to match all documents
	filter := bson.D{{}}

	// Delete all documents in the collection

   usermaxscoreCollection.DeleteMany(context.Background(), filter)
	
	testpaperCollection.DeleteMany(context.Background(),filter)
	totaluserCollection.DeleteMany(context.Background(), filter)
	//fmt.Println(result)
	httpClient.RespondSuccess(w,usermaxdata)
	
}
func DeleteTestInfo(w http.ResponseWriter, r *http.Request){

	var t struct {
		
		TestName string `json:"testname"`
		}
	
		err := json.NewDecoder(r.Body).Decode(&t)
		fmt.Println((t.TestName))
        if(err!=nil){
			fmt.Println(err)
		}
	// Create an empty filter to match all documents
	filter := bson.M{"name" :t.TestName}
   
		
	// Delete all documents in the collection
	result, err := testdetailsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
		
	}
	httpClient.RespondSuccess(w,result)
	
}
func LiveTestRank(testname string, User models.User)int{

	 number:=FindUserTotalNumber(testname,User)
	 //fmt.Println(number)
return FindNumberOFUserGreaterThen(testname,number)+1
}
func FindUserTotalNumber(testname string, User models.User)int {
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=usermaxscoreCollection.Find(ctx,bson.M{"testname":testname,"userphone":User.Phone,"testindex":-1})
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
	var user []UserMaxScore
if err =cursor.All(ctx,&user);err!=nil{
			fmt.Println(err)
		}
		//fmt.Println(user[0].TotalNumber)
	return user[0].TotalNumber

}

func FindNumberOFUserGreaterThen(testname string ,number int)int{

	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	
	cursor,err:=usermaxscoreCollection.Find(ctx,bson.M{
		"testname": testname,
		"testindex":-1,
		"totalnumber":   bson.M{"$gt": number},
	})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println(err)
		
	}
	
	

	type UserMaxScore struct {
	
	
		        UserData models.User `json:"user"`
				TestName string `json:"testname"`
				TestIndex int `json:"testindex"`
				TotalNumber int `json:"totalnumber"`
				
	
}
	var user []UserMaxScore
if err =cursor.All(ctx,&user);err!=nil{
			fmt.Println(err)
		}
	return len(user)
}
func ReturnAnswer(TestName string, index int) string{
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=questionCollection.Find(ctx,bson.M{"subject_tags":TestName})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println( "Please send a valid tag")
		
	}
	type QuestionType struct {
	
	
		ID             string        `json:"_id"`
		Correctanswer  string        `json:"correctanswer"`
		Correctanswers []string      `json:"correctanswers"`
		CreatedAt      time.Time     `json:"created_at"`
		Images         []interface{} `json:"images"`
		List1          interface{}   `json:"list1"`
		List2          interface{}   `json:"list2"`
		Options        []struct {
			Image string `json:"image"`
			Text  string `json:"text"`
		} `json:"options"`
		QID         string      `json:"q_id"`
		Question    string      `json:"question"`
		Solution    string      `json:"solution"`
		SubjectTags []string    `json:"subject_tags"`
		Type        string      `json:"type"`
		Usedby      interface{} `json:"usedby"`
	
}
 var questions []QuestionType
if err =cursor.All(ctx,&questions);err!=nil{
			fmt.Println( "Please send a valid tag")
			
		}
     if(questions[index].Correctanswer!=""){
		return questions[index].Correctanswer
	 }
	 return ArrayStringToString(questions[index].Correctanswers)
}
func GetNumberFromResponse(index int, resanswer string, testname string)int {
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=questionCollection.Find(ctx,bson.M{"subject_tags":testname})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println( "Please send a valid tag")
		
	}
	type QuestionType struct {
	
	
		ID             string        `json:"_id"`
		Correctanswer  string        `json:"correctanswer"`
		Correctanswers []string      `json:"correctanswers"`
		CreatedAt      time.Time     `json:"created_at"`
		Images         []interface{} `json:"images"`
		List1          interface{}   `json:"list1"`
		List2          interface{}   `json:"list2"`
		Options        []struct {
			Image string `json:"image"`
			Text  string `json:"text"`
		} `json:"options"`
		QID         string      `json:"q_id"`
		Question    string      `json:"question"`
		Solution    string      `json:"solution"`
		SubjectTags []string    `json:"subject_tags"`
		Type        string      `json:"type"`
		Usedby      interface{} `json:"usedby"`
	
}
 var questions []QuestionType
if err =cursor.All(ctx,&questions);err!=nil{
			fmt.Println( "Please send a valid tag")
			
		}
    qtype:=GetType(questions[index].Type)
    if(qtype==0){return GetNumberSingleCorrect(resanswer,questions[index].Correctanswer)}
	if(qtype==1){return GetNumberMultipleCorrect(resanswer,ArrayStringToString(questions[index].Correctanswers))}
	return GetNumberNumericalCorrect(resanswer,questions[index].Correctanswer)

}
func GetNumberSingleCorrect(answer string, correctanswer string)int {

	if(len(answer)==0){return 0}
	if(answer==correctanswer){return 3}
	return -1
}
func GetNumberNumericalCorrect(answer string, correctanswer string)int {

	if(len(answer)==0){return 0}
	a,_:=strconv.ParseFloat(answer,8)
	a=roundFloat(a, 2)
	b,_:=strconv.ParseFloat(correctanswer,8)
	b=roundFloat(b, 2)
	if(a==b){return 4}
	return 0
}
func GetNumberMultipleCorrect(answer string, correctanswer string)int {
  
   
	if(len(answer)==0){return 0}
	ans:=strings.Split(answer, ",")
	
	corans:=strings.Split(correctanswer,",")
	for i:=0;i<len(ans);i++{

		if(!strings.Contains(correctanswer,ans[i])){return -2}
	}
	if(len(ans)==len(corans)){return 4}
	return len(ans)

}
func roundFloat(val float64, precision uint) float64 {
    ratio := math.Pow(10, float64(precision))
    return math.Round(val*ratio) / ratio
}
func GetType(str string )int {
	if(str[0]=='S'){return 0}
	if(str[0]=='M'){return 1}
	if(str[0]=='N'){return 2}
	return 3;
}
func ArrayStringToString(str []string)string{

st:=str[0]
	for i := 1; i < len(str); i++ {
     st=st+","+str[i]
	}
	return st;
}
func MatchAnswer(answer string,correctanswer string)bool {
	if(answer==correctanswer){
		return true;
	}
return false

}
func LiveTestTimeValidation(name string) bool{
	indianTimeZone, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Error loading Indian time zone:", err)
		
	}
	indianTime := time.Now().In(indianTimeZone)
	s:=TestTime(name)
	//fmt.Println(s)
	start := time.Date(StringtoInt(s[0]), time.Month(StringtoInt(s[1])), StringtoInt(s[2]), StringtoInt(s[3]), StringtoInt(s[4]), StringtoInt(s[5]), 0,indianTimeZone)
	

	// Get the current time in the Indian time zone
	
	currenttime :=indianTime
	end := time.Date(StringtoInt(s[0])+3, time.Month(StringtoInt(s[1])), StringtoInt(s[2]), StringtoInt(s[3]), StringtoInt(s[4]), StringtoInt(s[5]), 0,indianTimeZone)
	if(currenttime.Compare(start)>=0 && currenttime.Compare(end)<=0){
		
		return true}
		
	return false;
	
	
}
func TestTime(testname string) []string{

	type PaperDetails struct {
		Name       string `json:"Name"`
		Start      string `json:"Start"`
		StartAt    string `json:"StartAt"`
		Difficulty string `json:"Difficulty"`
		Topics     string `json:"Topics"`
		Duration   string `json:"Duration"`
		Prize      string `json:"Prize"`
	}
	var ctx,_=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=testdetailsCollection.Find(ctx,bson.M{"name":testname})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println(err)
	}
	var papers []PaperDetails
	if err =cursor.All(ctx,&papers);err!=nil{
		fmt.Println(err)
	}

     //fmt.Println(papers[])
	 return strings.Split(papers[0].Start, "/")
	

//return strings.Split("2023/12/20/14/00/00","/")
}
func StringtoInt(s string)int {

	i,_:= strconv.Atoi(s)
	return i;
}
func LiveTestFindPaperIndex(name string) int{
	indianTimeZone, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Error loading Indian time zone:", err)
		
	}
	indianTime := time.Now().In(indianTimeZone)
    s:=TestTime(name)
	d := time.Date(StringtoInt(s[0]), time.Month(StringtoInt(s[1])), StringtoInt(s[2]), StringtoInt(s[3]), StringtoInt(s[4]), StringtoInt(s[5]), 0,indianTimeZone)
	currenttime := indianTime
	timegoes:=currenttime.Sub(d)
	
    maxquestions:=TotalQuestion(name)
	duration:=3*60*60/maxquestions
	timegoesinseconds:=timegoes.Hours()*60*60


 var index=int(timegoesinseconds)/(duration)
	return  index
}
func LiveTestFindQuestion(name string,index int){


}
func TotalQuestion(name string) int {
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=questionCollection.Find(ctx,bson.M{"subject_tags":name})
	defer cursor.Close(ctx)
	if err !=nil{

		fmt.Println(err )
		
	}
	var questions []bson.M
	if err =cursor.All(ctx,&questions);err!=nil{
		fmt.Println(err)
	}
	// fmt.Println(questions)
	return len(questions)

}
func GetRank(w http.ResponseWriter, r *http.Request){

	var t struct {
		Testname string `json:"testname"`
		User string `json:"user"`
		Number int `json:"number"`
	}


	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=testpaperCollection.Find(ctx,bson.M{"number":bson.M{"$gt":t.Number}})
    defer cursor.Close(ctx)
	if err !=nil{
                
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return

		
	}
	var users []bson.M
	if err =cursor.All(ctx,&users);err!=nil{
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}
	//cursor,err=testpaperCollection.Find(ctx,bson.M{"test_name":t.Testname,"User":t.User})
	
	httpClient.RespondSuccess(w, len(users)+1)

}
func InsertTestData(w http.ResponseWriter, r *http.Request ){

	var t struct {
		Testname string `json:"testname"`
		User string `json:"user"`
		Number int `json:"number"`
	}
	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second) 
		testpaperCollection.InsertOne(ctx,bson.M{"test_name":t.Testname,"user":t.User,"number":t.Number})

		httpClient.RespondSuccess(w, "success")
	
	

}
func UpdateTestData(w http.ResponseWriter, r *http.Request ){

	var t struct {
		Testname string `json:"testname"`
		User string `json:"user"`
		Number int `json:"number"`
	}
	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second) 
		testpaperCollection.UpdateOne(ctx,bson.M{"test_name":t.Testname,"user":t.User},bson.M{"$set":bson.M{"number":t.Number}})

		httpClient.RespondSuccess(w, "success")
	
	

}
