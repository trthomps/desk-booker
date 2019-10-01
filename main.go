package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/machinebox/graphql"
)

func main() {
	seatId := flag.String("seatId", os.Getenv("OS_SEAT_ID"), "Seat ID from OfficeSpace.  Default $OS_SEAT_ID.")
	employeeId := flag.String("employeeId", os.Getenv("OS_EMPLOYEE_ID"), "Employee ID from OfficeSpace.  Default $OS_EMPLOYEE_ID.")
	daysToBook := flag.Int("daysToBook", 10, "Number of days out to book, including today.")
	huddleSession := flag.String("huddleSession", os.Getenv("OS_HUDDLE_SESSION"), "OfficeSpace '_huddle_session' cookie.  Default $OS_HUDDLE_SESSION.")
	csrfToken := flag.String("csrfToken", os.Getenv("OS_CSRF_TOKEN"), "OfficeSpace 'X-CSRF-Token' header.  Default $OS_HUDDLE_SESSION.")

	flag.Parse()

	if *seatId == "" || *employeeId == "" || *huddleSession == "" || *csrfToken == "" {
		panic("seatId, employeeId, huddleSession, and csrfToken are required to be set.")
	}

	fmt.Printf("Attempting to book seat %s for %d days\n", *seatId, *daysToBook)

	client := graphql.NewClient("https://confluent.officespacesoftware.com/graphql")
	ctx := context.Background()

	for i := 0; i < *daysToBook; i++ {
		d := time.Now().Add(time.Duration(24*i) * time.Hour).Format("2006-01-02")

		fmt.Printf("Attempting to book day %d (%s)... ", i, d)

		req := graphql.NewRequest(CreateBookingQ)
		req.Var("input", map[string]interface{}{
			"seatId":       seatId,
			"employeeId":   employeeId,
			"checkInTime":  d,
			"checkOutTime": d,
		})

		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Cookie", "_huddle_session="+*huddleSession+"; webglv1=true; webglv2=true")
		req.Header.Set("X-CSRF-Token", *csrfToken)

		var respData map[string]interface{}
		if err := client.Run(ctx, req, &respData); err != nil {
			fmt.Println()
			panic(err)
		}

		cBooking := respData["createBooking"]
		if cBooking.(map[string]interface{})["error"] != nil {
			fmt.Printf("Failed. Error booking desk: %s\n", cBooking.(map[string]interface{})["error"].(map[string]interface{})["code"].(string))
			continue
		}

		booking := cBooking.(map[string]interface{})["booking"]
		id := booking.(map[string]interface{})["id"].(string)

		fmt.Printf("Success. [#%s] You should receive a booking confirmation email shortly.\n", id)
	}
}

const (
	CreateBookingQ = `
mutation createBooking($input: HotDeskBookingInputType!) {
  createBooking(input: $input) {
    affectedSeats {
      ...SeatInfo
      labelParts(
        withFreeAddressLabel: true
        withFreeAddressStatus: true
        settings: VISUAL_DIRECTORY
      ) {
        ...VisualDirectoryLabelPartsFragment
        __typename
      }
      employee {
        ...EmployeeResourceInfo
        __typename
      }
      __typename
    }
    booking {
      ...BookingInfo
      employee {
        fullName
        __typename
      }
      __typename
    }
    seat {
      ...SeatInfo
      labelParts(
        withFreeAddressLabel: true
        withFreeAddressStatus: true
        settings: VISUAL_DIRECTORY
      ) {
        ...VisualDirectoryLabelPartsFragment
        __typename
      }
      employee {
        ...EmployeeResourceInfo
        __typename
      }
      __typename
    }
    error {
      code
      conflictingBookings
      fieldName
      message
      __typename
    }
    __typename
  }
}

fragment EmployeeResourceInfo on Employee {
  id
  fullName
  department
  photo
  email
  seated: isSeated
  slackUsers {
    id
    slackTeamId
    slackUserId
    isPresent
    __typename
  }
  workPhone
  detailFields {
    href
    iconName
    label
    name
    type
    value
    __typename
  }
  __typename
}

fragment SeatInfo on Seat {
  availabilityDescription
  bookable
  description
  floorId
  iconName: visualDirectoryIconName
  id
  label
  managed
  openBookable
  shortDescription
  type
  utility
  vacant
  x
  y
  activeBooking {
    canceledAt
    canceledBy
    checkInBy
    checkInByEmail
    checkInTime
    checkOutBy
    checkOutScheduled
    checkOutTime
    createdAt
    employeeDepartment
    employeeEmail
    employeeFirstName
    employeeId
    employeeLastName
    employeeTitle
    floorId
    id
    localCheckOutScheduled
    roomDesignation
    seatId
    updatedAt
    updatedBy
    __typename
  }
  floor {
    id
    label
    __typename
  }
  halo {
    color
    floorId
    iconName
    id
    x
    y
    __typename
  }
  room {
    construct
    id
    points
    publicAttribute
    __typename
  }
  __typename
}

fragment BookingInfo on SeatOpenBooking {
  canceledAt
  canceledBy
  checkInTime
  checkOutBy
  checkOutScheduled
  checkOutTime
  createdAt
  endOfDayCheckOut
  employeeId
  employeeFirstName
  employeeLastName
  localTimeZone
  id
  key
  status
  __typename
}

fragment VisualDirectoryLabelPartsFragment on LabelPart {
  label
  occupants
  vacantLabel
  __typename
}`
)
