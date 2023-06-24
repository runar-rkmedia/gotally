package types

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jaswdr/faker"
)

func TestRules_Hash(t *testing.T) {
	t.Run("HashesShouldNotCollide", func(t *testing.T) {
		inputRules := map[string]Rules{}
		{
			// Using reflection, add a rules where only one field is set
			// add one more rule with the same field, but with a different value
			v := reflect.ValueOf(Rules{})
			typ := v.Type()

			for i := 0; i < v.NumField(); i++ {
				fake := faker.New()
				f := typ.Field(i)
				indirect := reflect.Indirect(v).FieldByName(f.Name).Interface()
				switch f.Name {
				case "ID", "CreatedAt", "UpdatedAt":
					continue
				default:
					// t.Error(f.Name, f.Type)
					rule1 := Rules{}
					rule2 := Rules{}
					switch indirect.(type) {
					case string:
						setStructValue(&rule1, f.Name, "foo")
						setStructValue(&rule2, f.Name, "bar")
					case time.Time:
						setStructValue(&rule1, f.Name, fake.Time().Time(time.Now()))
						setStructValue(&rule2, f.Name, fake.Time().Time(time.Now()))
					case uint64:
						setStructValue(&rule1, f.Name, uint64(1234))
						setStructValue(&rule2, f.Name, uint64(6789))
					case uint8:
						setStructValue(&rule1, f.Name, uint8(123))
						setStructValue(&rule2, f.Name, uint8(42))
					case bool:
						// Special case, since there are only two values here
						setStructValue(&rule1, f.Name, true)
						inputRules[f.Name+";"+v.Type().String()+"_a"] = rule1
						continue
					default:
						t.Fatalf("Unhandled type %s %v %#v ", f.Name, f.Type, v)
						continue

					}
					fmt.Println("set", f.Type.Kind(), f.Name, getStructValue(&rule1, f.Name), getStructValue(&rule2, f.Name))
					// t2 := v2.Type()
					// f2 := v2.FieldByName(f.Name)
					// fv2 := f2.Type()
					// x := reflect.NewAt(fv2, unsafe.Pointer(f2.UnsafeAddr()))
					// x.SetString("bobib")
					// fmt.Println("xx", &x)
					// if err := faker.FakeData(x); err != nil {
					// 	t.Fatal(err)
					// }
					inputRules[f.Name+";"+v.Type().String()+"_a"] = rule1
					inputRules[f.Name+";"+v.Type().String()+"_b"] = rule2
				}
			}

		}
		r := map[string]Rules{}
		for i, v := range inputRules {
			v := v
			hash := v.Hash()
			if v2, ok := r[hash]; ok {
				fieldName := strings.Split(i, ";")[0]
				val1 := getStructValue(&v, fieldName)
				val2 := getStructValue(&v2, fieldName)
				fmt.Println("Values for collisions", val1, val2)
				t.Errorf("Hash-collision detected at rule %s between values \n%#v\n%#v\nCheck that the types.Hash-method uses this field within its hash-generation\n", i, v, v2)
			}
			r[v.Hash()] = v
		}

	})
}

func getStructValue(arg interface{}, param string) any {

	v := reflect.ValueOf(arg).Elem()
	f := v.FieldByName(param)
	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return f.Uint()
	case reflect.String:
		return f.String()
	case reflect.Bool:
		return f.Bool()
	default:
		panic(fmt.Sprintf("Unhandled type %s during getStructValue", f.Kind()))
	}
}
func setStructValue(arg interface{}, param string, value interface{}) {
	v := reflect.ValueOf(arg).Elem()
	f := v.FieldByName(param)
	if !f.IsValid() {
		panic(fmt.Sprintf("reflect reported not valid on field %s in %#v", param, arg))
	}
	if !f.CanSet() {
		panic(fmt.Sprintf("reflect reported not settable on field %s in %#v", param, arg))
	}
	switch f.Kind() {
	case reflect.Int8:
		f.SetInt(int64(value.(int8)))
	case reflect.Uint8:
		f.SetUint(uint64(value.(uint8)))
	case reflect.Uint64:
		f.SetUint(value.(uint64))
	case reflect.String:
		f.SetString(value.(string))
	case reflect.Bool:
		f.SetBool(value.(bool))
	default:
		panic(fmt.Sprintf("Unhandled type %s during setStructValue", f.Kind()))
	}
}
