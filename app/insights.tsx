// type summary struct {
// 	Years []year
// }

// type year struct {
// 	Current  bool
// 	Contribs int
// 	Value    int
// }
import { BarChart, CartesianGrid, XAxis, YAxis, Tooltip, Legend, Bar, ResponsiveContainer } from "recharts"

interface Data {
    years: Array<year>
}

interface year {
    current: boolean
    contribs: number
    value: number
}

function maxYear(d: Data): year {
    return d.years.reduce(
        (accumulator, currentValue) => {
            if (accumulator.contribs < currentValue.contribs) {
                return currentValue
            }
            return accumulator
        }
    )
}

function toGraph(d: Data): Array<any> {
    return d.years.map((y) => ({
        name: y.value,
        value: y.contribs
    }))
}

interface InsightsProps {
    data: Data
}

// const App = ({ message }: AppProps)

const Insights = ({ data }: InsightsProps) => {
    console.log("Props are ", data)
    return <>
        <h2 className="text-2xl font-semibold mt-5">Insights</h2>
        <p className="mt-3">Here&apos;s what we can find out about you from your commits!</p>
        <h3 className="text-lg font-semibold mt-3">Yearly breakdown</h3>
        <p className="mt-3">The year with most contributions was {maxYear(data).value}. You made {maxYear(data).contribs} commits that year!</p>
        <h3 className="text-lg font-semibold mt-3">Yearly graph</h3>
        <ResponsiveContainer width="100%" height={250}>
            <BarChart margin={{ top: 20, right: 30, left: 0, bottom: 0 }} data={toGraph(data)}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" name="year" />
                <YAxis/>
                <Tooltip />
                <Legend />
                <Bar dataKey="value" name="contributions" fill="#8884d8" />
            </BarChart>
        </ResponsiveContainer>
    </>
}

export default Insights;