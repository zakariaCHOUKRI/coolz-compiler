declare i32 @printf(i8*, ...)

define i8* @Main_main(i8* %self) {
  %1 = call i32 (i8*, ...) @printf(i8* getelementptr inbounds ([3 x i8], [3 x i8]* @.fmt.ps, i32 0, i32 0), i8* getelementptr inbounds ([12 x i8], [12 x i8]* @.str.0, i32 0, i32 0))
ret i8* %self
}

@.str.0 = private unnamed_addr constant [12 x i8] c"Hello World\00"
@.fmt.ps = private unnamed_addr constant [4 x i8] c"%s\00\00"
define i32 @main() {
  %1 = call i8* @Main_main(i8* null)
  ret i32 0
}
