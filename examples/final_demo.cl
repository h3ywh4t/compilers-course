fun twice(x) {
    return x * 2;
}

var title = "result";
var nums = [1, 2, 3, 4];
var index = 1;
nums[index] = twice(10 + 5);

if (nums[1] == 30 && true) {
    print title + ": " + nums[1];
} else {
    print "bad";
}

while (false) {
    print "dead while body";
}

print [1, 2 + 3, 6];
