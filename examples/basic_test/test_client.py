import sys
from omniORB import CORBA
import Calculator

def main():
    if len(sys.argv) < 2:
        print("Usage: test_client.py <IOR>")
        sys.exit(1)
        
    ior = sys.argv[1]
    
    orb = CORBA.ORB_init(sys.argv, CORBA.ORB_ID)
    obj = orb.string_to_object(ior)
    
    math_obj = obj._narrow(Calculator.Math)
    
    if math_obj is None:
        print("Object reference is not a Calculator.Math")
        sys.exit(1)
        
    print("Invoking Add(10, 20)...")
    res1 = math_obj.Add(10, 20)
    print("Result:", res1)
    
    print("Invoking Echo('Hello from Python')...")
    res2 = math_obj.Echo("Hello from Python")
    print("Result:", res2)

if __name__ == "__main__":
    main()
