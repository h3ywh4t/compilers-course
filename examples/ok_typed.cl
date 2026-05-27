var title : string = "typed compiler";
var values : number[] = [1, 2, 3];

fun double(x : number) : number {
    return x * 2;
}

values[1] = double(10);
print title + ": " + values[1];
