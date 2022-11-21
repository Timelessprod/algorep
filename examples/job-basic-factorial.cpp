#include <iostream>
using namespace std;

int main()
{
    int n = 4;
    long factorial = 1;

    if (n < 0)
        cout << "Error! Factorial of a negative number doesn't exist.";
    else
    {
        for (int i = 1; i <= n; ++i)
        {
            factorial *= i;
        }
        cout << "Factorial of " << n << " = " << factorial;
    }

    return 0;
}